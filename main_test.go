package main

import "testing"

// genMockReader generates a mock reader function used to remove the need for a server
// list file to be present in the users home directory
func genMockReader(servers []string) serverListReader {
	mockReader := func(string) []string {
		return servers
	}
	return mockReader
}

func TestRunChecks(t *testing.T) {
	t.Log("Starting Test")

	mockReaderPass := genMockReader([]string{"google.com", "https://golang.org"})
	mockReaderFail := genMockReader([]string{"notfound.google.com", "https://golang.org"})

	testTable := []struct {
		mockReader     serverListReader
		expectedResult bool
	}{
		{mockReader: mockReaderPass, expectedResult: true},
		{mockReader: mockReaderFail, expectedResult: false},
	}

	for testnum, test := range testTable {

		serverList := newServerList(test.mockReader, "")

		serverList.read()

		t.Logf("Test Number:%v\n", testnum)
		t.Logf("Checking Servers:\n%v\n", serverList.sl)
		ok, got := runChecks(*serverList, 5)
		t.Logf("\nOk: %v\nGot:\n%v\n", ok, got)

		if test.expectedResult {
			if !ok {
				t.Errorf("Error checking servers:\n%v", got)
			}
			if got != "" {
				t.Errorf("Unexpected http response:\n%v", got)
			}
		} else if !test.expectedResult {
			if ok {
				t.Errorf("Check the servers in main_test.go, this should have failed but we got an ok:\n%v", got)
			}
		}
	}
}

func BenchmarkRunChecks(b *testing.B) {
	mockReader := genMockReader([]string{"google.com", "https://golang.org"})
	serverList := newServerList(mockReader, "")
	serverList.read()

	for n := 0; n < b.N; n++ {
		ok, got := runChecks(*serverList, 5)
		if !ok {
			b.Errorf("Sever check error during benchmarking:\n%v", got)
		}
	}
}
