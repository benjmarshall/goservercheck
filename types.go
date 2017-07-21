package main

// serverListReader is a function type used to allow mocking of the server list read in the unit tests
type serverListReader func(string) []string

// serverListType is a struct used to control generation/reading of a server list
type serverListType struct {
	reader         serverListReader
	serverListFile string
	sl             []string
}

// newServerList is the consructor method for the serverListType struct
func newServerList(reader serverListReader, serverListFile string) *serverListType {
	return &serverListType{reader: reader, serverListFile: serverListFile}
}

// read is a method on the serverListType which calls the reader function
func (serverList *serverListType) read() {
	serverList.sl = serverList.reader(serverList.serverListFile)
}

type serverResponse struct {
	url        string
	returnCode int
	status     string
}
