package snmp

import (
	"time"

	"github.com/gosnmp/gosnmp"
)

func GetSNMPInfo(ip string) (
	string,
	string,
	bool,
) {

	g := &gosnmp.GoSNMP{

		Target: ip,

		Port: 161,

		Community: "public",

		Version: gosnmp.Version2c,

		Timeout: time.Duration(2) *
			time.Second,

		Retries: 1,
	}

	err := g.Connect()

	if err != nil {

		return "", "", false
	}

	defer g.Conn.Close()

	oids := []string{

		"1.3.6.1.2.1.1.5.0",

		"1.3.6.1.2.1.1.1.0",
	}

	result, err := g.Get(oids)

	if err != nil {

		return "", "", false
	}

	systemName := ""

	systemDescription := ""

	for _, variable :=
		range result.Variables {

		if variable.Name ==
			".1.3.6.1.2.1.1.5.0" {

			systemName =
				string(variable.Value.([]byte))
		}

		if variable.Name ==
			".1.3.6.1.2.1.1.1.0" {

			systemDescription =
				string(variable.Value.([]byte))
		}
	}

	return systemName,
		systemDescription,
		true
}