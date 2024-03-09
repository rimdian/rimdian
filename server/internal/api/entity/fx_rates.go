package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/rotisserie/eris"
)

var (
	ErrFxRatesOutdated      = eris.New("fx rates outdated")
	ErrTargetFxRateNotFound = eris.New("target fx rate not found")

	round = func(x, unit float64) float64 {
		return float64(int64(x/unit+0.5)) * unit
	}
)

type FxRates struct {
	Base       string             `json:"base"`
	DateString string             `json:"date"`
	Rates      map[string]float64 `json:"rates"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

func (x *FxRates) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), &x)
}

func (x FxRates) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func (rates *FxRates) ConvertAmountToCurrency(fromCurrency string, toCurrency string, amount int64) (convertedAmount int64, error error) {
	// fmt.Printf("fromCurrency: %v", fromCurrency)
	// fmt.Printf("toCurrency: %v", toCurrency)
	// fmt.Printf("amount: %v", amount)

	if fromCurrency == toCurrency {
		return amount, nil
	}

	// do amount conversion
	// find destination currency

	// default rate if destination currency is EUR
	rate := 1.0

	for rateKey, value := range rates.Rates {
		if rateKey == toCurrency {
			rate = value
		}
	}

	// fmt.Printf("toCurrency currency rate: %v", rate)

	if fromCurrency != "EUR" {

		// convert to EUR first
		var fromRate float64

		for rateKey, value := range rates.Rates {
			if rateKey == fromCurrency {
				fromRate = value
			}
		}

		// fmt.Printf("fromCurrency currency rate: %v", fromRate)

		amount = int64(round(float64(amount)/fromRate, 0.05))

		// fmt.Printf("amount in EUR: %v", amount)
	}

	// if rate == 0 {
	//  return 0, errors.New(fmt.Sprintf("destination currency not found: %s", toCurrency))
	// }

	return int64(round(float64(amount)*rate, 0.05)), nil
}

// <?xml version="1.0" encoding="UTF-8"?>
// <gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
//  <gesmes:subject>Reference rates</gesmes:subject>
//  <gesmes:Sender>
//    <gesmes:name>European Central Bank</gesmes:name>
//  </gesmes:Sender>
//  <Cube>
//    <Cube time='2018-07-19'>
//      <Cube currency='USD' rate='1.1588'/>
//      <Cube currency='JPY' rate='130.98'/>
//      <Cube currency='BGN' rate='1.9558'/>
//      <Cube currency='CZK' rate='25.920'/>
//      <Cube currency='DKK' rate='7.4537'/>
//      <Cube currency='GBP' rate='0.89298'/>
//      <Cube currency='HUF' rate='325.77'/>
//      <Cube currency='PLN' rate='4.3280'/>
//      <Cube currency='RON' rate='4.6575'/>
//      <Cube currency='SEK' rate='10.3565'/>
//      <Cube currency='CHF' rate='1.1622'/>
//      <Cube currency='ISK' rate='124.40'/>
//      <Cube currency='NOK' rate='9.5763'/>
//      <Cube currency='HRK' rate='7.3938'/>
//      <Cube currency='RUB' rate='73.5585'/>
//      <Cube currency='TRY' rate='5.5957'/>
//      <Cube currency='AUD' rate='1.5804'/>
//      <Cube currency='BRL' rate='4.4874'/>
//      <Cube currency='CAD' rate='1.5351'/>
//      <Cube currency='CNY' rate='7.8553'/>
//      <Cube currency='HKD' rate='9.0963'/>
//      <Cube currency='IDR' rate='16773.63'/>
//      <Cube currency='ILS' rate='4.2393'/>
//      <Cube currency='INR' rate='80.0155'/>
//      <Cube currency='KRW' rate='1320.68'/>
//      <Cube currency='MXN' rate='22.1067'/>
//      <Cube currency='MYR' rate='4.7231'/>
//      <Cube currency='NZD' rate='1.7229'/>
//      <Cube currency='PHP' rate='62.180'/>
//      <Cube currency='SGD' rate='1.5909'/>
//      <Cube currency='THB' rate='38.797'/>
//      <Cube currency='ZAR' rate='15.6003'/>
//    </Cube>
//  </Cube>
// </gesmes:Envelope>

type FxCurrencyXML struct {
	Cube struct {
		Time  string `xml:"time,attr"`
		Items []struct {
			Currency string  `xml:"currency,attr"`
			Rate     float64 `xml:"rate,attr"`
		} `xml:"Cube"`
	} `xml:"Cube>Cube"`
}
