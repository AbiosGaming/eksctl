package delete

import (
	"github.com/kris-nova/logger"
	"github.com/spf13/pflag"

	api "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/authconfigmap"
	"github.com/weaveworks/eksctl/pkg/ctl/cmdutils"
	"github.com/weaveworks/eksctl/pkg/eks"
)

func deleteIAMIdentityMappingCmd(rc *cmdutils.ResourceCmd) {
	cfg := api.NewClusterConfig()
	rc.ClusterConfig = cfg

	var (
		arn authconfigmap.ARN
		all bool
	)

	rc.SetDescription("iamidentitymapping", "Delete a IAM identity mapping", "")

	rc.SetRunFunc(func() error {
		return doDeleteIAMIdentityMapping(rc, arn, all)
	})

	rc.FlagSetGroup.InFlagSet("General", func(fs *pflag.FlagSet) {
		fs.Var(&arn, "arn", "ARN of the IAM role or user to delete")
		fs.BoolVar(&all, "all", false, "Delete all matching mappings instead of just one")
		cmdutils.AddNameFlag(fs, cfg.Metadata)
		cmdutils.AddRegionFlag(fs, rc.ProviderConfig)
		cmdutils.AddConfigFileFlag(fs, &rc.ClusterConfigFile)
	})

	cmdutils.AddCommonFlagsForAWS(rc.FlagSetGroup, rc.ProviderConfig, false)
}

func doDeleteIAMIdentityMapping(rc *cmdutils.ResourceCmd, arn authconfigmap.ARN, all bool) error {
	if err := cmdutils.NewMetadataLoader(rc).Load(); err != nil {
		return err
	}

	cfg := rc.ClusterConfig

	ctl := eks.New(rc.ProviderConfig, cfg)

	if err := ctl.CheckAuth(); err != nil {
		return err
	}

	if arn.Resource == "" {
		return cmdutils.ErrMustBeSet("--arn")
	}
	if cfg.Metadata.Name == "" {
		return cmdutils.ErrMustBeSet("--name")
	}

	if err := ctl.GetCredentials(cfg); err != nil {
		return err
	}
	clientSet, err := ctl.NewStdClientSet(cfg)
	if err != nil {
		return err
	}
	acm, err := authconfigmap.NewFromClientSet(clientSet)
	if err != nil {
		return err
	}

	if err := acm.RemoveIdentity(arn, all); err != nil {
		return err
	}
	if err := acm.Save(); err != nil {
		return err
	}

	// Check whether we have more roles that match
	identities, err := acm.Identities()
	if err != nil {
		return err
	}
	filtered := identities.Get(arn)
	if len(filtered) > 0 {
		logger.Warning("there are %d mappings left with same arn %q (use --all to delete them at once)", len(filtered), arn)
	}
	return nil
}
