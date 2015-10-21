package hmacs


import (
	"testing"
	"os"
	"encoding/csv"
	"fmt"
)

type Test struct {
	test *testing.T
}

type HmacTestCase struct {
	EncryptionKey string
	Sample string
	HMAC string
}
func (this *Test) GetHmacTestCase(algorithm string) (testCases []*HmacTestCase) {
	testCases, err := this.getHmacTestCase(algorithm)
	if err != nil {
		this.test.Errorf("failed to get test cases for algorithm '%s': %s", algorithm, err.Error())
	}
	return
}

func (this *Test) AssertEquals(message string, expected, actual string) {
	if actual != expected {
		this.test.Errorf(message + ": expected '%v' is different from actual '%v'  ", expected, actual)
	}
}

func (this *Test) getHmacTestCase(algorithm string) (testCases []*HmacTestCase, err error) {
	file, err := os.Open("hmacSamples." + algorithm + ".csv")
	if err != nil { return }
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '	'
	reader.Comment = '#'
	reader.FieldsPerRecord = 3

	samples, err := reader.ReadAll()
	if err != nil { return }

	testCases = make([]*HmacTestCase, len(samples));
	for index, sample := range samples {
		testCases[index] = &HmacTestCase{
			HMAC: sample[0],
			EncryptionKey: sample[1],
			Sample: sample[2],
		}
	}
	return
}

func TestStartAndStop(t *testing.T) {
	test := Test{t}
	testCases := test.GetHmacTestCase("sha1")

	for _, testCase := range testCases {
		generator := HashHmacGenerator{[]byte(testCase.EncryptionKey)}
		actual := generator.Sha1String(testCase.Sample)
		test.AssertEquals(
			fmt.Sprintf("incorrect hmac for encryption key '%s' and sample '%s'", testCase.EncryptionKey, testCase.Sample),
			testCase.HMAC,
			actual,
		)
	}
}