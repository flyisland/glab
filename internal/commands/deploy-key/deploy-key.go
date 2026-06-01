package deploykey

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"gitlab.com/gitlab-org/cli/internal/cmdutils"
	cmdAdd "gitlab.com/gitlab-org/cli/internal/commands/deploy-key/add"
	cmdDelete "gitlab.com/gitlab-org/cli/internal/commands/deploy-key/delete"
	cmdGet "gitlab.com/gitlab-org/cli/internal/commands/deploy-key/get"
	cmdList "gitlab.com/gitlab-org/cli/internal/commands/deploy-key/list"
)

func NewCmdDeployKey(f cmdutils.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-key <command>",
		Short: "Manage deploy keys.",
		Long: heredoc.Docf(`
			Add, list, get, and delete the deploy keys for a project.

			Deploy keys grant access to a repository over SSH without being tied to a
			user account, and are commonly used by CI/CD jobs and external systems.
			These commands operate on the current project. Use %[1]s--repo%[1]s to target
			another project.
		`, "`"),
	}

	cmdutils.EnableRepoOverride(cmd, f)

	cmd.AddCommand(cmdAdd.NewCmdAdd(f))
	cmd.AddCommand(cmdGet.NewCmdGet(f))
	cmd.AddCommand(cmdList.NewCmdList(f))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f))

	return cmd
}
