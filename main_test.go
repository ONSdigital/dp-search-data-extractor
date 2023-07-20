package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-data-extractor/features/steps"
	dplogs "github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	t *testing.T
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	component := steps.NewComponent(f.t)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if err := component.Reset(); err != nil {
			return ctx, fmt.Errorf("unable to initialise scenario: %s", err)
		}
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		component.Close()
		return ctx, nil
	})

	component.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(_ *godog.TestSuiteContext) {
	dplogs.Namespace = "dp-search-data-exporter"
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		f := &ComponentTest{t: t}

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
