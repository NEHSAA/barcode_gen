package memberdb

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/NEHSAA/barcode_gen/common/log"

	sheets "google.golang.org/api/sheets/v4"
)

type MemberDbController struct {
	client        *http.Client
	logger        log.Logger
	spreadsheetID string
	sheetTitle    string

	sheetSvc *sheets.Service
}

type Member struct {
	SN              string `parse:"#"`
	Name            string `parse:"姓名"`
	Gender          string `parse:"性別"`
	NationalID      string `parse:"身分證字號"`
	MembershipType  string `parse:"會員種類"`
	Lifelong        string `parse:"永久會員"`
	MembershipYears string `parse:"會員資格"`
	IDCardStatus    string `parse:"會員證狀態"`
	IDCardCount     string `parse:"會員證補發"`
}

func GetMemberDbController(ctx context.Context, client *http.Client,
	spreadsheetID, sheetTitle string) (*MemberDbController, error) {

	logger := log.GetLogrusLogger("memberdb")
	srv, err := sheets.New(client)
	if err != nil {
		logger.Errorf("cannot retrieve sheet client: %s", err)
		return nil, fmt.Errorf("cannot retrieve sheet client: %s", err)
	}

	// fields := []googleapi.Field{"#", "姓名", "性別", "身分證字號", "會員種類", "永久會員", "會員資格", "會員證狀態", "會員證補發"}
	// fields := []googleapi.Field{"Title"}

	return &MemberDbController{client, logger, spreadsheetID, sheetTitle, srv}, nil
}

func (c *MemberDbController) getAllRawData() ([][]interface{}, error) {
	c.logger.Infof("fetching spreadsheet id %v", c.spreadsheetID)
	// spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Fields().Do()
	title, err := c.sheetSvc.Spreadsheets.Values.Get(c.spreadsheetID, fmt.Sprintf("'%s'!1:2000", c.sheetTitle)).Do()
	if err != nil {
		c.logger.Errorf("Unable to retrieve data: %v", err)
		return nil, fmt.Errorf("Unable to retrieve data: %v", err)
	}

	return title.Values, nil
}

func reformData(header, data []interface{}) (map[string]string, error) {
	numEntries := len(header)
	if len(data) < numEntries {
		numEntries = len(data)
	}
	result := make(map[string]string, numEntries)
	for i := 0; i < numEntries; i++ {
		title, ok := header[i].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert header column index %v", i)
		}
		value, ok := data[i].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert data column index %v", i)
		}
		result[title] = value
	}
	return result, nil
}

const parseTagName = "parse"

var parseHeaderToFieldOnce sync.Once
var parseHeaderToField map[string]string

func buildParseHeaderToField() {
	parseHeaderToFieldOnce.Do(
		func() {
			ret := Member{}
			t := reflect.TypeOf(ret)
			parseHeaderToField = make(map[string]string, t.NumField())
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				tag := field.Tag.Get(parseTagName)
				parseHeaderToField[tag] = field.Name
			}
		},
	)
}

func newMember(data map[string]string) (*Member, error) {
	ret := Member{}
	if parseHeaderToField == nil {
		buildParseHeaderToField()
	}

	r := reflect.ValueOf(&ret)
	for k, v := range data {
		field, ok := parseHeaderToField[k]
		if !ok {
			continue
		}
		reflect.Indirect(r).FieldByName(field).SetString(v)
	}

	return &ret, nil
}

func (c *MemberDbController) GetAllData() ([]Member, error) {
	data, err := c.getAllRawData()
	if err != nil {
		return nil, err
	}
	c.logger.Infof("fetched %v rows", len(data))

	if len(data) <= 1 {
		return []Member{}, nil
	}

	ret := make([]Member, len(data)-1)

	header := data[0]
	for i, row := range data {
		if i == 0 {
			continue
		}
		// c.logger.Debugf("processing %v", row)
		reformed, err := reformData(header, row)
		if err != nil {
			return nil, err
		}
		// c.logger.Debugf("reformed %v", reformed)
		parsed, err := newMember(reformed)
		if err != nil {
			return nil, err
		}
		ret[i-1] = *parsed
	}
	return ret, nil
}

func (c *MemberDbController) GetById(id string) (int, *Member, error) {
	allData, err := c.GetAllData()
	if err != nil {
		return 0, nil, err
	}
	for i, v := range allData {
		if v.NationalID == id {
			return i, &v, nil
		}
	}
	return 0, nil, nil
}

func (c *MemberDbController) SetIDCardStatus(id, status string) (*Member, error) {
	data, err := c.getAllRawData()
	if err != nil {
		return nil, err
	}
	c.logger.Infof("fetched %v rows", len(data))

	if len(data) <= 1 {
		return nil, nil
	}

	targetRow := -1
	var targetMember *Member = nil

	header := data[0]
	for i, row := range data {
		if i == 0 {
			continue
		}
		// c.logger.Debugf("processing %v", row)
		reformed, err := reformData(header, row)
		if err != nil {
			return nil, err
		}
		// c.logger.Debugf("reformed %v", reformed)
		parsed, err := newMember(reformed)
		if err != nil {
			return nil, err
		}
		if parsed.NationalID == id {
			targetRow = i + 1
			targetMember = parsed
			break
		}
	}
	if targetRow == -1 {
		return nil, fmt.Errorf("member (target row) not found")
	}

	targetCol := ""
	for i, v := range header {
		if v.(string) == "會員證狀態" {
			targetCol = string('A' + i)
		}
	}
	if targetCol == "" {
		return nil, fmt.Errorf("target column not found")
	}

	var vr sheets.ValueRange
	// myval := []interface{}{status}
	// vr.Values = append(vr.Values, myval)
	// vr.Range = fmt.Sprintf("%s%d", targetCol, targetRow)
	vr.Values = [][]interface{}{[]interface{}{status}}

	// fmt.Println(fmt.Sprintf("%s%d", targetCol, targetRow), vr)
	_, err = c.sheetSvc.Spreadsheets.Values.Update(
		c.spreadsheetID,
		fmt.Sprintf("%s!%s%d", c.sheetTitle, targetCol, targetRow),
		&vr).ValueInputOption("RAW").Do()

	if err != nil {
		return nil, err
	}
	targetMember.IDCardStatus = status
	return targetMember, nil
}

func getBool(s string) (bool, error) {
	senitized := strings.ToLower(strings.TrimSpace(s))
	if senitized == "true" || senitized == "yes" || senitized == "1" {
		return true, nil
	} else if senitized == "false" || senitized == "no" || senitized == "0" {
		return false, nil
	}
	return false, fmt.Errorf("failed to parse \"%v\" as bool", s)
}

func (m Member) GetMembershipTypeForBarcode() (string, error) {
	parser := regexp.MustCompile("^.*會員")
	membershipType := parser.FindStringSubmatch(m.MembershipType)[0]

	if membershipType != "一般會員" {
		return membershipType, nil
	}
	lifelong, err := getBool(m.Lifelong)
	if err != nil {
		return "", fmt.Errorf("failed to determine if it's lifelong")
	}
	if lifelong {
		return "永久會員", nil
	}
	return "一般會員", nil
}
