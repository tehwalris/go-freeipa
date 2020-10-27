package freeipa

import "io"

type KerberosConnectOptions struct {
	Krb5ConfigReader io.Reader
	KeytabReader     io.Reader
	Username         string
	Realm            string
}
