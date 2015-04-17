package diego

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/pivotal-cf-experimental/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("Security Groups", func() {
	type DoraCurlResponse struct {
		Stdout     string
		Stderr     string
		ReturnCode int `json:"return_code"`
	}

	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	// this test assumes the default running security groups block access to the DEAs
	// the test takes advantage of the fact that the DEA ip address and internal container ip address
	//  are discoverable via the cc api and dora's myip endpoint
	It("allows traffic and then blocks traffic", func() {
		By("pushing it")
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Dora, "--no-start", "-b", "ruby_buildpack"), CF_PUSH_TIMEOUT).Should(Exit(0))

		By("staging and running it on Diego")
		enableDiego(appName)
		Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))

		By("verifying it's up")
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Dora!"))

		secureAddress := helpers.LoadConfig().SecureAddress
		secureHost, securePort, err := net.SplitHostPort(secureAddress)
		Ω(err).ShouldNot(HaveOccurred())

		// test app egress rules
		var doraCurlResponse DoraCurlResponse
		curlResponse := helpers.CurlApp(appName, fmt.Sprintf("/curl/%s/%s", secureHost, securePort))
		json.Unmarshal([]byte(curlResponse), &doraCurlResponse)
		Expect(doraCurlResponse.ReturnCode).ShouldNot(Equal(0))
		firstCurlError := doraCurlResponse.ReturnCode

		// apply security group
		rules := fmt.Sprintf(`[{"destination":"%s","ports":"%s","protocol":"tcp"}]`, secureHost, securePort)

		file, _ := ioutil.TempFile(os.TempDir(), "DATS-sg-rules")
		defer os.Remove(file.Name())
		file.WriteString(rules)
		file.Close()

		rulesPath := file.Name()
		securityGroupName := fmt.Sprintf("DATS-SG-%s", generator.RandomName())

		cf.AsUser(context.AdminUserContext(), time.Minute, func() {
			Eventually(cf.Cf("create-security-group", securityGroupName, rulesPath)).Should(Exit(0))
			Eventually(
				cf.Cf("bind-security-group",
					securityGroupName,
					context.RegularUserContext().Org,
					context.RegularUserContext().Space)).Should(Exit(0))
		})
		defer func() {
			cf.AsUser(context.AdminUserContext(), time.Minute, func() {
				Eventually(cf.Cf("delete-security-group", securityGroupName, "-f")).Should(Exit(0))
			})
		}()

		Eventually(cf.Cf("restart", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Dora!"))

		// test app egress rules
		curlResponse = helpers.CurlApp(appName, fmt.Sprintf("/curl/%s/%s", secureHost, securePort))
		json.Unmarshal([]byte(curlResponse), &doraCurlResponse)
		Expect(doraCurlResponse.ReturnCode).Should(Equal(0))

		// unapply security group
		cf.AsUser(context.AdminUserContext(), time.Minute, func() {
			Eventually(
				cf.Cf("unbind-security-group",
					securityGroupName, context.RegularUserContext().Org,
					context.RegularUserContext().Space)).
				Should(Exit(0))
		})

		By("restarting it - without security group")
		Eventually(cf.Cf("restart", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Dora!"))

		// test app egress rules
		curlResponse = helpers.CurlApp(appName, fmt.Sprintf("/curl/%s/%s", secureHost, securePort))
		json.Unmarshal([]byte(curlResponse), &doraCurlResponse)
		Expect(doraCurlResponse.ReturnCode).Should(Equal(firstCurlError))
	})
})
