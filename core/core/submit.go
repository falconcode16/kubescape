package core

import (
	"github.com/armosec/kubescape/core/cautils"
	"github.com/armosec/kubescape/core/cautils/getter"
	"github.com/armosec/kubescape/core/cautils/logger"
	"github.com/armosec/kubescape/core/cautils/logger/helpers"
	"github.com/armosec/kubescape/core/meta/cliinterfaces"
)

func (ks *Kubescape) Submit(submitInterfaces cliinterfaces.SubmitInterfaces) error {

	// list resources
	postureReport, err := submitInterfaces.SubmitObjects.SetResourcesReport()
	if err != nil {
		return err
	}
	allresources, err := submitInterfaces.SubmitObjects.ListAllResources()
	if err != nil {
		return err
	}
	// report
	if err := submitInterfaces.Reporter.Submit(&cautils.OPASessionObj{PostureReport: postureReport, AllResources: allresources}); err != nil {
		return err
	}
	logger.L().Success("Data has been submitted successfully")
	submitInterfaces.Reporter.DisplayReportURL()

	return nil
}

func (ks *Kubescape) SubmitExceptions(accountID, excPath string) error {
	logger.L().Info("submitting exceptions", helpers.String("path", excPath))

	// load cached config
	tenantConfig := getTenantConfig(accountID, "", getKubernetesApi())
	if err := tenantConfig.SetTenant(); err != nil {
		logger.L().Error("failed setting account ID", helpers.Error(err))
	}

	// load exceptions from file
	loader := getter.NewLoadPolicy([]string{excPath})
	exceptions, err := loader.GetExceptions("")
	if err != nil {
		return err
	}

	// login kubescape SaaS
	armoAPI := getter.GetArmoAPIConnector()
	if err := armoAPI.Login(); err != nil {
		return err
	}

	if err := armoAPI.PostExceptions(exceptions); err != nil {
		return err
	}
	logger.L().Success("Exceptions submitted successfully")

	return nil
}
