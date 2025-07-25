package filter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

type RegexpFilter struct {
	Value    *regexp.Regexp
	valueRaw string
	bodyOnly bool
}

func NewRegexpFilter(value string, bodyOnly bool) (ffuf.FilterProvider, error) {
	re, err := regexp.Compile(value)
	if err != nil {
		return &RegexpFilter{}, fmt.Errorf("Regexp filter or matcher (-fr / -mr): invalid value: %s", value)
	}
	return &RegexpFilter{Value: re, valueRaw: value, bodyOnly: bodyOnly}, nil
}

func (f *RegexpFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Value string `json:"value"`
	}{
		Value: f.valueRaw,
	})
}

func (f *RegexpFilter) Filter(response *ffuf.Response) (bool, error) {
	var matchdata []byte
	if f.bodyOnly {
		matchdata = response.Data
	} else {
		matchheaders := ""
		for k, v := range response.Headers {
			for _, iv := range v {
				matchheaders += k + ": " + iv + "\r\n"
			}
		}
		matchdata = []byte(matchheaders)
		matchdata = append(matchdata, response.Data...)
	}

	pattern := f.valueRaw
	for keyword, inputitem := range response.Request.Input {
		pattern = strings.ReplaceAll(pattern, keyword, regexp.QuoteMeta(string(inputitem)))
	}
	matched, err := regexp.Match(pattern, matchdata)
	if err != nil {
		return false, nil
	}
	return matched, nil
}

func (f *RegexpFilter) Repr() string {
	return f.valueRaw
}

func (f *RegexpFilter) ReprVerbose() string {
	return fmt.Sprintf("Regexp: %s", f.valueRaw)
}
