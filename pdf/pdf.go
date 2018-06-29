package pdf

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"

	"common/log"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/signintech/gopdf"
)

const (
	dpi  float64 = 72.0
	mm   float64 = dpi / 25.4
	inch float64 = dpi
)

const (
	TextSeparator = "•"
)

var (
	pdfDefaultInfo = gopdf.PdfInfo{
		Title:   "nehsaa-barcode",
		Author:  "nehsaa",
		Subject: "nehsaa-barcode",
		Creator: "gopdf",
	}
)

type IdBarcodeData struct {
	Text           string
	BarcodeContent string
}

func GetIdBarcodePdf(ctx context.Context, data []IdBarcodeData) ([]byte, error) {
	logger := log.GetLogger(ctx)
	logger.Infof("generating pdf based on data: %v", data)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 3.2 * inch, H: 9 * mm}})
	pdf.SetInfo(pdfDefaultInfo)

	err := pdf.AddTTFFont("main", "DFYuan-W5-WIN-BF-01.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to add font: %v", err)
	}

	for _, entry := range data {
		logger.Infof("generating page for entry: %v", entry)
		pdf.AddPage()

		err = pdf.SetFont("main", "", 11)
		if err != nil {
			return nil, err
		}
		pdf.SetX(1.0 * mm)
		pdf.SetY(2.5 * mm)
		pdf.Cell(nil, entry.Text)

		var b barcode.Barcode
		b, err = code128.Encode(entry.BarcodeContent)
		if err != nil {
			return nil, fmt.Errorf("failed to generate barcode: %v", err)
		}
		barcodeW := (3.2-1.5)*inch - 3*mm
		barcodeH := 7 * mm
		b, err = barcode.Scale(b, int(barcodeW), int(barcodeH))
		if err != nil {
			return nil, fmt.Errorf("failed to scale barcode: %v", err)
		}
		var bbuf bytes.Buffer
		err = jpeg.Encode(&bbuf, b, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to encode barcode to png: %v", err)
		}

		bimg, err := gopdf.ImageHolderByReader(&bbuf)
		if err != nil {
			return nil, fmt.Errorf("failed to create img holder: %v", err)
		}
		err = pdf.ImageByHolder(bimg, 1.5*inch, 1*mm, &gopdf.Rect{H: barcodeH, W: barcodeW})
		if err != nil {
			return nil, fmt.Errorf("failed to place img holder: %v", err)
		}
	}

	return pdf.GetBytesPdfReturnErr()
}
