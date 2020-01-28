package commands_test

import (
	"io/ioutil"
	"net/http"
	"os"

	"runtime"

	"code.cloudfoundry.org/credhub-cli/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

func withTemporaryFile(wantingFile func(string)) error {
	f, err := ioutil.TempFile("", "credhub_tests_")

	if err != nil {
		return err
	}

	name := f.Name()

	f.Close()
	wantingFile(name)

	return os.Remove(name)
}

var _ = Describe("Export", func() {
	BeforeEach(func() {
		login()
	})

	ItRequiresAuthentication("export")
	ItRequiresAnAPIToBeSet("export")

	testAutoLogIns := []TestAutoLogin{
		{
			method:              "GET",
			responseFixtureFile: "get_response.json",
			responseStatus:      http.StatusOK,
			endpoint:            "/api/v1/data",
		},
		{
			method:              "GET",
			responseFixtureFile: "get_certs_response.json",
			responseStatus:      http.StatusOK,
			endpoint:            "/api/v1/certificates/",
		},
	}
	ItAutomaticallyLogsIn(testAutoLogIns, "export")

	ItBehavesLikeHelp("export", "e", func(session *Session) {
		Expect(session.Err).To(Say("Usage"))
		if runtime.GOOS == "windows" {
			Expect(session.Err).To(Say("credhub-cli.exe \\[OPTIONS\\] export \\[export-OPTIONS\\]"))
		} else {
			Expect(session.Err).To(Say("credhub-cli \\[OPTIONS\\] export \\[export-OPTIONS\\]"))
		}
	})

	Describe("Exporting", func() {
		It("queries for the most recent version of all credentials", func() {
			findJson := `{
				"credentials": [
					{
						"version_created_at": "idc",
						"name": "/path/to/cred"
					},
					{
						"version_created_at": "idc",
						"name": "/path/to/another/cred"
					}
				]
			}`

			getJson := `{
				"data": [{
					"type":"value",
					"id":"some_uuid",
					"name":"/path/to/cred",
					"version_created_at":"idc",
					"value": "foo"
				}]
			}`

			certificates := `{
				"certificates": [{
					"id": "cert-id",
					"name": "/cert",
					"signed_by": "/cert_ca",
					"signs": [],
					"versions": [{
							"certificate_authority": false,
							"expiry_date": "2020-11-28T14:04:40Z",
							"generated": true,
							"id": "cert-version-id",
							"self_signed": false,
							"transitional": false
						}
					]}, {
					"id": "cert-ca-id",
					"name": "/cert_ca",
					"signed_by": "/cert_ca",
					"signs": [
						"/cert"
					],
					"versions": [{
							"certificate_authority": true,
							"expiry_date": "2020-11-28T14:04:38Z",
							"generated": true,
							"id": "cert-ca-version-id",
							"self_signed": true,
							"transitional": false
						}
					]}
				]}`

			responseTable := `credentials:
- name: /path/to/cred
  type: value
  value: foo
- name: /path/to/cred
  type: value
  value: foo`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/data", "path="),
					RespondWith(http.StatusOK, findJson),
				),
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/certificates/"),
					RespondWith(http.StatusOK, certificates),
				),
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/data", "name=/path/to/cred&current=true"),
					RespondWith(http.StatusOK, getJson),
				),
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/data", "name=/path/to/another/cred&current=true"),
					RespondWith(http.StatusOK, getJson),
				),
			)

			session := runCommand("export")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(responseTable))
		})

		Context("when given a path", func() {
			It("queries for credentials matching that path", func() {
				noCredsJson := `{ "credentials" : [] }`
				noCertificates := `{ "certificates" : [] }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/api/v1/data", "path=some/path"),
						RespondWith(http.StatusOK, noCredsJson),
					),
					CombineHandlers(
						VerifyRequest("GET", "/api/v1/certificates/"),
						RespondWith(http.StatusOK, noCertificates),
					),
				)

				session := runCommand("export", "-p", "some/path")

				Eventually(session).Should(Exit(0))
			})
		})

		Context("when given a file", func() {
			It("writes the YAML to that file", func() {
				withTemporaryFile(func(filename string) {
					noCredsJson := `{ "credentials" : [] }`
					noCertificates := `{ "certificates" : [] }`
					noCredsYaml := `credentials: []
`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/api/v1/data", "path="),
							RespondWith(http.StatusOK, noCredsJson),
						),
						CombineHandlers(
							VerifyRequest("GET", "/api/v1/certificates/"),
							RespondWith(http.StatusOK, noCertificates),
						),
					)

					session := runCommand("export", "-f", filename)

					Eventually(session).Should(Exit(0))

					Expect(filename).To(BeAnExistingFile())

					fileContents, _ := ioutil.ReadFile(filename)

					Expect(string(fileContents)).To(Equal(noCredsYaml))
				})
			})
		})
	})

	Describe("Errors", func() {
		It("prints an error when the network request fails", func() {
			cfg := config.ReadConfig()
			cfg.ApiURL = "mashed://potatoes"
			config.WriteConfig(cfg)

			session := runCommand("export")

			Eventually(session).Should(Exit(1))
			Eventually(string(session.Err.Contents())).Should(ContainSubstring("Get mashed://potatoes/api/v1/data?path=: unsupported protocol scheme \"mashed\""))
		})

		It("prints an error if the specified output file cannot be opened", func() {
			noCredsJson := `{ "credentials" : [] }`
			noCertificates := `{ "certificates" : [] }`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/data", "path="),
					RespondWith(http.StatusOK, noCredsJson),
				),
				CombineHandlers(
					VerifyRequest("GET", "/api/v1/certificates/"),
					RespondWith(http.StatusOK, noCertificates),
				),
			)

			session := runCommand("export", "-f", "this/should/not/exist/anywhere")

			Eventually(session).Should(Exit(1))
			if runtime.GOOS == "windows" {
				Eventually(string(session.Err.Contents())).Should(ContainSubstring("open this/should/not/exist/anywhere: The system cannot find the path specified"))
			} else {
				Eventually(string(session.Err.Contents())).Should(ContainSubstring("open this/should/not/exist/anywhere: no such file or directory"))
			}
		})
	})
})
