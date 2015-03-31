package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Docker Registry", func() {
	var appName string
	var createDockerAppPayload string

	domain := helpers.LoadConfig().AppsDomain

	type AppSummary struct {
		DockerImage string `json:"docker_image"`
	}

	BeforeEach(func() {
		appName = generator.RandomName()

		createDockerAppPayload = `{
			"name": "%s",
			"memory": 512,
			"instances": 1,
			"disk_quota": 1024,
			"space_guid": "%s",
			"docker_image": "cloudfoundry/diego-docker-app:latest",
			"command": "/myapp/dockerapp",
			"diego": true
		}`
	})

	JustBeforeEach(func() {
		spaceGuid := guidForSpaceName(context.RegularUserContext().Space)

		payload := fmt.Sprintf(createDockerAppPayload, appName, spaceGuid)
		Eventually(cf.Cf("curl", "/v2/apps", "-X", "POST", "-d", payload)).Should(Exit(0))
		Eventually(cf.Cf("create-route", context.RegularUserContext().Space, domain, "-n", appName)).Should(Exit(0))
		Eventually(cf.Cf("map-route", appName, domain, "-n", appName)).Should(Exit(0))
		Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
		Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Describe("running the app with private registry", func() {
		var imageName string
		var address string

		JustBeforeEach(func() {
			appGuid := guidForAppName(appName)

			appSummary := cf.Cf("curl", fmt.Sprintf("/v2/apps/%s/summary", appGuid))
			Ω(appSummary.Wait()).To(Exit(0))

			var appData AppSummary
			err := json.Unmarshal(appSummary.Out.Contents(), &appData)
			Ω(err).ShouldNot(HaveOccurred())

			slashIndex := strings.Index(appData.DockerImage, "/")
			Ω(slashIndex).ShouldNot(Equal(-1))
			tagIndex := strings.LastIndex(appData.DockerImage, ":")
			Ω(tagIndex).ShouldNot(Equal(-1))

			address = appData.DockerImage[0:slashIndex]
			imageName = appData.DockerImage[slashIndex+1 : tagIndex]
		})

		It("stores the public image in the private registry", func() {
			client := http.Client{}
			resp, err := client.Get(fmt.Sprintf("http://%s/v1/search?q=%s", address, imageName))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp.StatusCode).Should(Equal(http.StatusOK))
			bytes, err := ioutil.ReadAll(resp.Body)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(bytes)).Should(ContainSubstring("library/" + imageName))
		})
	})
})
