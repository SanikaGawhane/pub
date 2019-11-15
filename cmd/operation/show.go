package operation

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/devigned/pub/pkg/service"

	"github.com/devigned/pub/cmd/args"
	"github.com/devigned/pub/pkg/partner"
	"github.com/devigned/pub/pkg/xcobra"
)

type (
	showOperationsArgs struct {
		Publisher string
		Offer     string
		Operation string
	}
)

func newShowCommand(sl service.CommandServicer) (*cobra.Command, error) {
	var oArgs showOperationsArgs
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show an operation by Id",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) {
			client, err := sl.GetCloudPartnerService()
			if err != nil {
				log.Fatalf("unable to create Cloud Partner Portal client: %v", err)
			}

			op, err := client.GetOperation(ctx, partner.GetOperationParams{
				PublisherID: oArgs.Publisher,
				OfferID:     oArgs.Offer,
				OperationID: oArgs.Operation,
			})

			if err != nil {
				xcobra.PrintfErrAndExit(1, "unable to fetch operations: %v", err)
			}

			if err := sl.GetPrinter().Print(op); err != nil {
				log.Fatalf("unable to print operation: %v", err)
			}
		}),
	}

	if err := args.BindPublisher(cmd, &oArgs.Publisher); err != nil {
		return cmd, err
	}

	if err := args.BindOffer(cmd, &oArgs.Offer); err != nil {
		return cmd, err
	}

	cmd.Flags().StringVar(&oArgs.Operation, "op", "", "Operation Id (guid).")
	err := cmd.MarkFlagRequired("op")
	return cmd, err
}
