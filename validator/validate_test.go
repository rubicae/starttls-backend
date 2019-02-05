package validator

import (
	"testing"

	"time"

	"github.com/EFForg/starttls-backend/checker"
)

type mockDomainPolicyStore struct {
	hostnames map[string][]string
}

func (m mockDomainPolicyStore) DomainsToValidate() ([]string, error) {
	domains := []string{}
	for domain := range m.hostnames {
		domains = append(domains, domain)
	}
	return domains, nil
}

func (m mockDomainPolicyStore) HostnamesForDomain(domain string) ([]string, error) {
	return m.hostnames[domain], nil
}

func mockValidate(domainMap map[string][]string, check performCheckFunc, report reportFailureFunc) {
	mockStore := mockDomainPolicyStore{hostnames: domainMap}
	validator := Validator{
		Name: "mock", Interval: 100 * time.Millisecond, Store: mockStore,
		checkPerformer: check, FailureReporter: report}
	validator.Start()
}

func TestRegularValidationValidates(t *testing.T) {
	called := make(chan bool)
	fakeChecker := func(domain string, hostnames []string) checker.DomainResult {
		called <- true
		return checker.DomainResult{}
	}
	go mockValidate(map[string][]string{"a": []string{"hostname"}}, fakeChecker, nil)

	select {
	case <-called:
		return
	case <-time.After(time.Second):
		t.Errorf("Checker wasn't called on hostname!")
	}
}

func TestRegularValidationReportsErrors(t *testing.T) {
	reports := make(chan string)
	fakeChecker := func(domain string, hostnames []string) checker.DomainResult {
		if domain == "fail" || domain == "error" {
			return checker.DomainResult{Status: 5}
		}
		return checker.DomainResult{Status: 0}
	}
	fakeReporter := func(name string, domain string, result checker.DomainResult) {
		reports <- domain
	}
	go mockValidate(map[string][]string{
		"fail":   []string{"hostname"},
		"error":  []string{"hostname"},
		"normal": []string{"hostname"}}, fakeChecker, fakeReporter)
	recvd := make(map[string]bool)
	numRecvd := 0
	for numRecvd < 4 {
		select {
		case report := <-reports:
			recvd[report] = true
			numRecvd++
		case <-time.After(time.Second):
			t.Errorf("Timed out waiting for reports")
		}
	}
	if _, ok := recvd["fail"]; !ok {
		t.Errorf("Expected fail to be reported")
	}
	if _, ok := recvd["error"]; !ok {
		t.Errorf("Expected error to be reported")
	}
	if _, ok := recvd["normal"]; ok {
		t.Errorf("Didn't expected normal to be reported")
	}
}
