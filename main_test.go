package main

import (
	"flag"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-data-extractor/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	// KafkaFeature *componenttest.KafkaFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {

	testComponent := steps.NewComponent()

	ctx.BeforeScenario(func(*godog.Scenario) {
		testComponent.Reset()
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		testComponent.Close()
	})

	testComponent.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
