package version

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/devigned/pub/pkg/partner"
	"github.com/devigned/pub/pkg/xcobra"
)

func init() {
	listCmd.Flags().StringVarP(&listVersionsArgs.Publisher, "publisher", "p", "", "publisher ID for your Cloud Partner Provider")
	_ = listCmd.MarkFlagRequired("publisher")
	listCmd.Flags().StringVarP(&listVersionsArgs.Offer, "offer", "o", "", "String that uniquely identifies the offer (Offer ID).")
	_ = listCmd.MarkFlagRequired("offer")
	listCmd.Flags().StringVarP(&listVersionsArgs.SKU, "sku", "s", "", "String that uniquely identifies the SKU (SKU ID).")
	_ = listCmd.MarkFlagRequired("sku")
	rootCmd.AddCommand(listCmd)
}

type (
	// ListVersionsArgs are the arguments for `versions list` command
	ListVersionsArgs struct {
		Publisher string
		Offer     string
		SKU       string
	}
)

var (
	listVersionsArgs ListVersionsArgs
	listCmd          = &cobra.Command{
		Use:   "list",
		Short: "list all versions for a given plan",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) {
			client, err := getClient()
			if err != nil {
				log.Fatalf("unable to create Cloud Partner Portal client: %v", err)
			}

			offer, err := client.GetOffer(ctx, partner.ShowOfferParams{
				PublisherID: listVersionsArgs.Publisher,
				OfferID:     listVersionsArgs.Offer,
			})

			if err != nil {
				log.Fatalf("unable to list offers: %v", err)
			}

			var versions map[string]partner.VirtualMachineImage
			for _, plan := range offer.Definition.Plans {
				if plan.ID == listVersionsArgs.SKU {
					versions = plan.GetVMImages()
					break
				}
			}

			printVersions(versions)
		}),
	}
)

func printVersions(versions map[string]partner.VirtualMachineImage) {
	bits, err := json.Marshal(versions)
	if err != nil {
		log.Fatalf("failed to print plans: %v", err)
	}
	fmt.Print(string(bits))
}