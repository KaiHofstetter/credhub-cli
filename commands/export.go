package commands

import (
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/credhub-cli/config"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"code.cloudfoundry.org/credhub-cli/models"
)

type ExportCommand struct {
	Path string `short:"p" long:"path" description:"Path of credentials to export" required:"false"`
	File string `short:"f" long:"file" description:"File in which to write credentials" required:"false"`
}

func (cmd ExportCommand) Execute([]string) error {
	allCredentials, err := getAllCredentialsForPath(cmd.Path)

	if err != nil {
		return err
	}

	exportCreds, err := models.ExportCredentials(allCredentials)

	if err != nil {
		return err
	}

	if cmd.File == "" {
		fmt.Printf("%s", exportCreds)

		return err
	} else {
		return ioutil.WriteFile(cmd.File, exportCreds.Bytes, 0644)
	}
}

func getAllCredentialsForPath(path string) ([]credentials.Credential, error) {
	cfg := config.ReadConfig()
	credhubClient, err := initializeCredhubClient(cfg)

	if err != nil {
		return nil, err
	}

	allPaths, err := credhubClient.FindByPath(path)

	if err != nil {
		return nil, err
	}

	allCertsMetadata, err := credhubClient.GetAllCertificatesMetadata()

	if err != nil {
		return nil, err
	}

	credentials := make([]credentials.Credential, len(allPaths.Credentials))
	for i, baseCred := range allPaths.Credentials {
		credential, err := credhubClient.GetLatestVersion(baseCred.Name)

		if err != nil {
			return nil, err
		}

		var value interface{}

		if cert, ok := credential.Value.(values.Certificate); ok {
			if caName, caNameOk := extractCaName(credential.Name, allCertsMetadata); caNameOk {
				cert.CaName = caName
				value = cert
			}
		} else {
			value = credential.Value
		}
		credential.Value = value
		credentials[i] = credential
	}

	return credentials, nil
}

func extractCaName(certName string, certificates []credentials.CertificateMetadata) (string, bool) {
	for _, certMetadata := range certificates {
		for _, sign := range certMetadata.Signs {
			if sign == certName {
				return certMetadata.Name, true
			}
		}
	}
	return "", false
}