package diego

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/pivotal-cf-experimental/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("Application Lifecycle", func() {
	var appName string

	apps := func() *Session {
		return cf.Cf("apps").Wait()
	}

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	reportedIDs := func(instances int) map[string]bool {
		seenIDs := map[string]bool{}
		for len(seenIDs) != instances {
			seenIDs[helpers.CurlApp(appName, "/id")] = true
		}

		return seenIDs
	}

	differentIDsFrom := func(idsBefore map[string]bool) []string {
		differentIDs := []string{}

		for id, _ := range reportedIDs(len(idsBefore)) {
			if !idsBefore[id] {
				differentIDs = append(differentIDs, id)
			}
		}

		return differentIDs
	}

	Describe("An app staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it")
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Dora, "--no-start", "-b", "ruby_buildpack"), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("staging and running it on Diego")
			enableDiego(appName)
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("verifying it's up")
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Dora!"))

			By("checking its LANG")
			Ω(helpers.CurlApp(appName, "/env/LANG")).Should(ContainSubstring("en_US.UTF-8"))

			By("verifying the buildpack's detect never runs")
			appGuid := guidForAppName(appName)
			Eventually(cf.Cf("curl", "/v2/apps/"+appGuid)).Should(Say(`"detected_buildpack": ""`))

			By("stopping it")
			Eventually(cf.Cf("stop", appName)).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))

			By("setting an environment variable")
			Eventually(cf.Cf("set-env", appName, "LANG", "en_GB.ISO8859-1")).Should(Exit(0))

			By("starting it")
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Dora!"))

			By("checking its LANG")
			Ω(helpers.CurlApp(appName, "/env/LANG")).Should(ContainSubstring("en_GB.ISO8859-1"))

			By("scaling it")
			Eventually(cf.Cf("scale", appName, "-i", "2")).Should(Exit(0))
			Eventually(apps).Should(Say("2/2"))

			idsBefore := reportedIDs(2)

			By("restarting an instance")
			Eventually(cf.Cf("restart-app-instance", appName, "1")).Should(Exit(0))
			Eventually(func() []string {
				return differentIDsFrom(idsBefore)
			}).Should(HaveLen(1))

			idsBefore = reportedIDs(2)

			By("recovering from crashes")
			helpers.CurlApp(appName, "/sigterm/KILL")
			Eventually(func() []string {
				return differentIDsFrom(idsBefore)
			}, 10*time.Second).Should(HaveLen(1))
		})

		It("being reported as 'crashed' after enough crashes", func() {
			By("pushing it")
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Dora, "-c", "/bin/false", "--no-start", "-b", "ruby_buildpack"), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("staging and running it on Diego")
			enableDiego(appName)
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(1))

			Eventually(cf.Cf("app", appName)).Should(Say("crashed"))
			Eventually(cf.Cf("events", appName)).Should(Say("app.crash.*exit_description:"))
		})
	})
})
