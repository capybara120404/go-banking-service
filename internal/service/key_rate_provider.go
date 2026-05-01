package service

import (
	"bytes"
	"fmt"
	"go-banking-service/internal/logger"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
)

type CentralBankKeyRateProvider struct {
	client *http.Client
}

func NewCentralBankKeyRateProvider() *CentralBankKeyRateProvider {
	return &CentralBankKeyRateProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *CentralBankKeyRateProvider) GetKeyRate() (float64, error) {
	soapRequest := p.buildSOAPRequest()
	rawBody, err := p.sendRequest(soapRequest)
	if err != nil {
		logger.Error("Failed to send request to CBR", "error", err)
		return 0, err
	}
	rate, err := p.parseXMLResponse(rawBody)
	if err != nil {
		logger.Error("Failed to parse XML response from CBR", "error", err)
		return 0, err
	}

	rate += 5
	logger.Info("Retrieved key rate from CBR", "rate", rate)
	return rate, nil
}

func (p *CentralBankKeyRateProvider) buildSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
        <soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
            <soap12:Body>
                <KeyRate xmlns="http://web.cbr.ru/">
                    <fromDate>%s</fromDate>
                    <ToDate>%s</ToDate>
                </KeyRate>
            </soap12:Body>
        </soap12:Envelope>`, fromDate, toDate)
}

func (p *CentralBankKeyRateProvider) sendRequest(soapRequest string) ([]byte, error) {
	req, err := http.NewRequest(
		"POST",
		"https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	return rawBody, nil
}

func (p *CentralBankKeyRateProvider) parseXMLResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("ошибка парсинга XML: %v", err)
	}

	krElements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(krElements) == 0 {
		return 0, fmt.Errorf("данные по ставке не найдены")
	}
	latestKR := krElements[0]
	rateElement := latestKR.FindElement("./Rate")
	if rateElement == nil {
		return 0, fmt.Errorf("тег Rate отсутствует")
	}

	rateStr := rateElement.Text()
	var rate float64
	if _, err := fmt.Sscanf(rateStr, "%f", &rate); err != nil {
		return 0, fmt.Errorf("ошибка конвертации ставки: %v", err)
	}

	return rate, nil
}
