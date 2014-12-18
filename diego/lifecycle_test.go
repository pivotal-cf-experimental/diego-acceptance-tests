package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/diego-acceptance-tests/helpers/assets"
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

	Describe("An app staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it")
			Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", ZIP_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("staging and running it on Diego")
			Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("verifying it's up")
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))

			By("stopping it")
			Eventually(cf.Cf("stop", appName)).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))

			By("starting it")
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))

			By("scaling it")
			Eventually(cf.Cf("scale", appName, "-i", "2")).Should(Exit(0))
			Eventually(apps).Should(Say("2/2"))

			By("restarting an instance")
			Eventually(cf.Cf("restart-app-instance", appName, "1")).Should(Exit(0))
			Eventually(apps).Should(Say("1/2"))
			Eventually(apps).Should(Say("2/2"))
		})
	})
})
