package http_server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NEHSAA/barcode_gen/common/log"
	"github.com/NEHSAA/barcode_gen/memberdb"
	"github.com/NEHSAA/barcode_gen/pdf"

	"github.com/gorilla/mux"
)

func (s *Server) handleTestEmail(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	client, err := s.getGoogleClient(ctx, r)
	if err != nil {
		if err == ErrorGAuthTokenDoesNotExist || err == ErrorGAuthTokenExpired {
			s.redirectToGAuthLogin(w, r, r.RequestURI)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to load gauth client: %s", err)))
		return
	}

	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("cannot retrieve user info: %s", err)))
		return
	}

	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	w.Write([]byte(fmt.Sprintf("Email body: %v", string(data))))
}

func (s *Server) handleGetSpreadsheet(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	client, err := s.getGoogleClient(ctx, r)
	if err != nil {
		if err == ErrorGAuthTokenDoesNotExist || err == ErrorGAuthTokenExpired {
			s.redirectToGAuthLogin(w, r, r.RequestURI)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to load gauth client: %s", err)))
		return
	}

	vars := mux.Vars(r)
	spreadsheetId := vars["spreadsheet_id"]
	sheetTitle := vars["sheet_title"]

	ctrl, err := memberdb.GetMemberDbController(ctx, client, spreadsheetId, sheetTitle)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Unable to retrieve sheet: %v", err)))
		w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}

	data, err := ctrl.GetAllData()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Unable to get header: %v", err)))
		w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}
	// w.Write([]byte(fmt.Sprintln(hdr)))

	get_tr := func(m memberdb.Member) string {
		ret := "<tr>"
		ret += fmt.Sprintf("<td>%v</td>", m.SN)
		ret += fmt.Sprintf("<td>%v</td>", m.Name)
		ret += fmt.Sprintf("<td>%v</td>", m.Gender)
		ret += fmt.Sprintf("<td>%v</td>", m.NationalID)
		ret += fmt.Sprintf("<td>%v</td>", m.MembershipType)
		ret += fmt.Sprintf("<td>%v</td>", m.Lifelong)
		ret += fmt.Sprintf("<td>%v</td>", m.MembershipYears)
		ret += fmt.Sprintf("<td>%v</td>", m.IDCardStatus)
		ret += fmt.Sprintf("<td>%v</td>", m.IDCardCount)
		ret += fmt.Sprintf("<td><a href=\"/makepdf/%s/%s/%s\">barcode</a></td>", m.Name, m.MembershipType, m.NationalID)
		ret += "</tr>"
		return ret
	}

	w.Write([]byte("<html><body>"))
	w.Write([]byte("<table><tr><th>#</th> <th>姓名</th> <th>性別</th> <th>身分證字號</th> <th>會員種類</th> <th>永久會員</th> <th>會員資格</th> <th>會員證狀態</th> <th>會員證補發</th> <th></th> </tr>"))
	for _, m := range data {
		w.Write([]byte(get_tr(m)))
	}
	w.Write([]byte("</html></body>"))
}

func (s *Server) handleVerifyIDCard(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	client, err := s.getGoogleClient(ctx, r)
	if err != nil {
		if err == ErrorGAuthTokenDoesNotExist || err == ErrorGAuthTokenExpired {
			s.redirectToGAuthLogin(w, r, r.RequestURI)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to load gauth client: %s", err)))
		return
	}

	vars := mux.Vars(r)
	spreadsheetId := vars["spreadsheet_id"]
	sheetTitle := vars["sheet_title"]

	ctrl, err := memberdb.GetMemberDbController(ctx, client, spreadsheetId, sheetTitle)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Unable to retrieve sheet: %v", err)))
		w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}

	getHtml := func(msg string) string {
		return `
<html>
<meta charset="UTF-8">
<body>
<form method="get">
  請刷製作完畢的條碼: <input type="text" name="id" autofocus>
  <input type="submit" value="Submit">
</form>
<pre>
` + msg + `
</pre>
</body></html>`
	}

	logger := log.GetLogrusLogger("verfy")
	ids, ok := r.URL.Query()["id"]
	if !ok {
		logger.Debugf("nothing to verify for: %v", ids)
		w.Write([]byte(getHtml("")))
		return
	}
	msg := ""

	logger.Infof("verifying: %v", ids)
	msg = fmt.Sprintf("verifying: %v\n", ids)

	for _, id := range ids {
		msg += fmt.Sprintf("\n==> %v\n", id)
		// rid, member, err := ctrl.GetById(id)
		// if err != nil {
		// 	msg += fmt.Sprintf("錯誤: %v\n", err)
		// 	continue
		// }
		// if member == nil {
		// 	msg += fmt.Sprintf("錯誤: 查無會員/身分證字號錯誤\n")
		// 	continue
		// }
		// msg += fmt.Sprintf("查到會員: %v (row #%v)\n", member.Name, rid)

		member, err := ctrl.SetIDCardStatus(id, "produced")
		if err != nil {
			msg += fmt.Sprintf("更改狀態錯誤: %v\n", err)
			continue
		}
		msg += fmt.Sprintf("%s: 更改狀態成功\n", member.Name)
	}

	w.Write([]byte(getHtml(msg)))

	// data, err := ctrl.GetAllData()
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte("<html><body>"))
	// 	w.Write([]byte(fmt.Sprintf("Unable to get header: %v", err)))
	// 	w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
	// 	w.Write([]byte("</html></body>"))
	// 	return
	// }
}

func (s *Server) handleGetPdf(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	client, err := s.getGoogleClient(ctx, r)
	if err != nil {
		if err == ErrorGAuthTokenDoesNotExist || err == ErrorGAuthTokenExpired {
			s.redirectToGAuthLogin(w, r, r.RequestURI)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to load gauth client: %s", err)))
		return
	}

	vars := mux.Vars(r)
	spreadsheetId := vars["spreadsheet_id"]
	sheetTitle := vars["sheet_title"]
	memberNationalId := vars["id"]

	ctrl, err := memberdb.GetMemberDbController(ctx, client, spreadsheetId, sheetTitle)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Unable to retrieve sheet: %v", err)))
		w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}

	_, member, err := ctrl.GetById(memberNationalId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("錯誤: %v\n", err)))
		return
	}
	if member == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("錯誤: 查無會員/身分證字號錯誤\n")))
		return
	}

	label, err := member.GetMembershipTypeForBarcode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("資料格式錯誤: %v\n", err)))
		return
	}

	data := []pdf.IdBarcodeData{
		{member.Name + pdf.TextSeparator + label,
			member.NationalID},
	}
	pdf, err := pdf.GetIdBarcodePdf(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	content := bytes.NewReader(pdf)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename='nehsaa_barcode.pdf")
	http.ServeContent(w, r, "pdf", time.Now(), content)
}

func (s *Server) handleGetAllPendingPdf(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	client, err := s.getGoogleClient(ctx, r)
	if err != nil {
		if err == ErrorGAuthTokenDoesNotExist || err == ErrorGAuthTokenExpired {
			s.redirectToGAuthLogin(w, r, r.RequestURI)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to load gauth client: %s", err)))
		return
	}

	vars := mux.Vars(r)
	spreadsheetId := vars["spreadsheet_id"]
	sheetTitle := vars["sheet_title"]

	ctrl, err := memberdb.GetMemberDbController(ctx, client, spreadsheetId, sheetTitle)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Unable to retrieve sheet: %v", err)))
		w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}

	memberData, err := ctrl.GetAllData()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Unable to get header: %v", err)))
		w.Write([]byte("<p><a href=\"/auth/login?redirect=" + r.RequestURI + "\">retry login here</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}

	barcodeData := []pdf.IdBarcodeData{}
	for _, member := range memberData {
		if member.IDCardStatus != "pending" {
			continue
		}

		label, err := member.GetMembershipTypeForBarcode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("資料格式錯誤: %v\n", err)))
			return
		}

		barcodeData = append(barcodeData,
			pdf.IdBarcodeData{member.Name + pdf.TextSeparator + label,
				member.NationalID})
	}

	pdf, err := pdf.GetIdBarcodePdf(barcodeData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	content := bytes.NewReader(pdf)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename='nehsaa_barcode.pdf")
	http.ServeContent(w, r, "pdf", time.Now(), content)
}

func (s *Server) mountMemberSpreadsheetEndpoints() {
	s.Router.HandleFunc("/member/testEmail", s.handleTestEmail)
	s.Router.HandleFunc("/member/list/spreadsheet/{spreadsheet_id}/sheet/{sheet_title}", s.handleGetSpreadsheet)
	s.Router.HandleFunc("/member/verify/spreadsheet/{spreadsheet_id}/sheet/{sheet_title}", s.handleVerifyIDCard)
	s.Router.HandleFunc("/member/getPdf/spreadsheet/{spreadsheet_id}/sheet/{sheet_title}/{id}", s.handleGetPdf)
	s.Router.HandleFunc("/member/getAllPendingPdf/spreadsheet/{spreadsheet_id}/sheet/{sheet_title}", s.handleGetAllPendingPdf)
}
