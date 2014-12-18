package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/diego-acceptance-tests/helpers/assets"
)

var _ = Describe("DEA Compatibility", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
		Eventually(cf.Cf("push", appName, "-p", assets.NewAssets().Standalone, "--no-start", "-b", ZIP_NULL_BUILDPACK), CF_PUSH_TIMEOUT).Should(Exit(0))
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Describe("An app staged with Diego and running on a DEA", func() {
		BeforeEach(func() {
			Eventually(cf.Cf("set-env", appName, DIEGO_STAGE_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
		})

		It("comes up", func() {
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
		})
	})

	Describe("An app staged on the DEA and running on Diego", func() {
		BeforeEach(func() {
			Eventually(cf.Cf("set-env", appName, DIEGO_RUN_BETA, "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
		})

		It("comes up", func() {
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Bash!"))
		})
	})
})
