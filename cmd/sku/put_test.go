package sku

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/devigned/pub/internal/test"
	"github.com/devigned/pub/pkg/partner"
)

func TestPutCommand_FailOnInsufficientArgs(t *testing.T) {
	test.VerifyFailsOnArgs(t, newPutCommand)
	test.VerifyFailsOnArgs(t, newPutCommand, "-p", "foo")
	test.VerifyFailsOnArgs(t, newPutCommand, "-p", "foo", "-o", "bar")
}

func TestPutCommand_FailOnCloudPartnerError(t *testing.T) {
	_, fName, del := test.NewTmpSKUFile(t, "sku", "sku_one", "skuSummary")
	defer del()

	test.VerifyCloudPartnerServiceCommand(t, newPutCommand, "-p", "foo", "-o", "bar", "-f", fName)
}

func TestPutCommand_FailOnGetOfferError(t *testing.T) {
	_, skuFileName, del := test.NewTmpSKUFile(t, "sku", "sku_one", "skuSummary")
	defer del()

	boomErr := errors.New("boom")
	svcMock := new(test.CloudPartnerServiceMock)
	svcMock.
		On("GetOffer", mock.Anything, partner.ShowOfferParams{
			PublisherID: "foo",
			OfferID:     "bar",
		}).
		Return(&partner.Offer{}, boomErr)

	prtMock := new(test.PrinterMock)
	prtMock.
		On("ErrPrintf", "unable to get offer: %v", []interface{}{boomErr}).
		Return(nil)

	rm := new(test.RegistryMock)
	rm.On("GetCloudPartnerService").Return(svcMock, nil)
	rm.On("GetPrinter").Return(prtMock)

	cmd, err := test.QuietCommand(newPutCommand(rm))
	require.NoError(t, err)
	cmd.SetArgs([]string{"-p", "foo", "-o", "bar", "-f", skuFileName})
	assert.Error(t, cmd.Execute())
	prtMock.AssertCalled(t, "ErrPrintf", "unable to get offer: %v", []interface{}{boomErr})
}

func TestPutCommand_WarnIfPlanAlreadyExistsForOffer(t *testing.T) {
	_, skuFileName, del := test.NewTmpSKUFile(t, "sku", "planId_one", "New summary")
	defer del()

	offer := test.NewMarketplaceVMOffer()
	assert.Equal(t, 1, len(offer.Definition.Plans))
	assert.Equal(t, "planId_one", offer.Definition.Plans[0].ID)

	svcMock := new(test.CloudPartnerServiceMock)
	svcMock.On("GetOffer", mock.Anything, mock.Anything).Return(offer, nil)

	warning := "Plan 'planId_one' already exists for offer 'test'"

	prtMock := new(test.PrinterMock)
	prtMock.On("Print", warning).Return(nil)

	rm := new(test.RegistryMock)
	rm.On("GetCloudPartnerService").Return(svcMock, nil)
	rm.On("GetPrinter").Return(prtMock)

	cmd, err := test.QuietCommand(newPutCommand(rm))
	require.NoError(t, err)
	cmd.SetArgs([]string{"-p", offer.PublisherID, "-o", offer.ID, "-f", skuFileName})
	assert.NoError(t, cmd.Execute())
	prtMock.AssertCalled(t, "Print", warning)
}

func TestPutCommand_ForceReplacePlanIfAlreadyExistsForOffer(t *testing.T) {
	plan, skuFileName, del := test.NewTmpSKUFile(t, "sku", "planId_one", "New summary")
	defer del()

	offer := test.NewMarketplaceVMOffer()
	assert.Equal(t, 1, len(offer.Definition.Plans))
	assert.Equal(t, "planId_one", offer.Definition.Plans[0].ID)

	expectedOffer := test.NewMarketplaceVMOffer()
	expectedOffer.Definition.Plans = []partner.Plan{plan}
	assert.Equal(t, 1, len(expectedOffer.Definition.Plans))
	assert.Equal(t, "planId_one", expectedOffer.Definition.Plans[0].ID)
	assert.Equal(t, "New summary", expectedOffer.Definition.Plans[0].PlanVirtualMachineDetail.SKUSummary)

	svcMock := new(test.CloudPartnerServiceMock)
	svcMock.On("GetOffer", mock.Anything, mock.Anything).Return(offer, nil)
	svcMock.On("PutOffer", mock.Anything, expectedOffer).Return(expectedOffer, nil)

	prtMock := new(test.PrinterMock)
	prtMock.On("Print", expectedOffer).Return(nil)

	rm := new(test.RegistryMock)
	rm.On("GetCloudPartnerService").Return(svcMock, nil)
	rm.On("GetPrinter").Return(prtMock)

	cmd, err := test.QuietCommand(newPutCommand(rm))
	require.NoError(t, err)
	cmd.SetArgs([]string{"--force", "-p", offer.PublisherID, "-o", offer.ID, "-f", skuFileName})
	assert.NoError(t, cmd.Execute())
	prtMock.AssertCalled(t, "Print", expectedOffer)
}

func TestPutCommand_FailOnPutOfferError(t *testing.T) {
	offer := test.NewMarketplaceVMOffer()
	_, skuFileName, del := test.NewTmpSKUFile(t, "sku", "new_sku", "skuSummary")
	defer del()

	boomErr := errors.New("boom")

	svcMock := new(test.CloudPartnerServiceMock)
	svcMock.
		On("GetOffer", mock.Anything, partner.ShowOfferParams{
			PublisherID: offer.PublisherID,
			OfferID:     offer.ID,
		}).
		Return(offer, nil)
	svcMock.
		On("PutOffer", mock.Anything, mock.Anything).
		Return(offer, boomErr)

	prtMock := new(test.PrinterMock)
	prtMock.
		On("ErrPrintf", "unable to put offer: %v", []interface{}{boomErr}).
		Return(nil)

	rm := new(test.RegistryMock)
	rm.On("GetCloudPartnerService").Return(svcMock, nil)
	rm.On("GetPrinter").Return(prtMock)

	cmd, err := test.QuietCommand(newPutCommand(rm))
	require.NoError(t, err)
	cmd.SetArgs([]string{"-p", offer.PublisherID, "-o", offer.ID, "-f", skuFileName})
	assert.Error(t, cmd.Execute())
	prtMock.AssertCalled(t, "ErrPrintf", "unable to put offer: %v", []interface{}{boomErr})
}

func TestPutCommand_SuccessPutFirstPlan(t *testing.T) {
	plan, skuFileName, del := test.NewTmpSKUFile(t, "sku", "first_sku", "skuSummary")
	defer del()

	offer := test.NewMarketplaceVMOffer()
	offer.Definition.Plans = nil
	assert.Equal(t, 0, len(offer.Definition.Plans))

	expectedOffer := test.NewMarketplaceVMOffer()
	expectedOffer.Definition.Plans = nil
	expectedOffer.Definition.Plans = append(expectedOffer.Definition.Plans, plan)

	assert.Equal(t, 1, len(expectedOffer.Definition.Plans))
	assert.Equal(t, "first_sku", expectedOffer.Definition.Plans[0].ID)

	svcMock := new(test.CloudPartnerServiceMock)
	svcMock.On("GetOffer", mock.Anything, mock.Anything).Return(offer, nil)
	svcMock.On("PutOffer", mock.Anything, expectedOffer).Return(expectedOffer, nil)

	prtMock := new(test.PrinterMock)
	prtMock.On("Print", expectedOffer).Return(nil)

	rm := new(test.RegistryMock)
	rm.On("GetCloudPartnerService").Return(svcMock, nil)
	rm.On("GetPrinter").Return(prtMock)

	cmd, err := test.QuietCommand(newPutCommand(rm))
	require.NoError(t, err)
	cmd.SetArgs([]string{"-p", offer.PublisherID, "-o", offer.ID, "-f", skuFileName})
	assert.NoError(t, cmd.Execute())
	prtMock.AssertCalled(t, "Print", expectedOffer)
}

func TestPutCommand_SuccessPutAdditionalPlan(t *testing.T) {
	plan, skuFileName, del := test.NewTmpSKUFile(t, "sku", "second_sku", "skuSummary")
	defer del()

	offer := test.NewMarketplaceVMOffer()
	assert.Equal(t, 1, len(offer.Definition.Plans))
	assert.NotEqual(t, "second_sku", offer.Definition.Plans[0].ID)

	expectedOffer := test.NewMarketplaceVMOffer()
	expectedOffer.Definition.Plans = append(expectedOffer.Definition.Plans, plan)
	assert.Equal(t, 2, len(expectedOffer.Definition.Plans))
	assert.Equal(t, "second_sku", expectedOffer.Definition.Plans[1].ID)

	svcMock := new(test.CloudPartnerServiceMock)
	svcMock.On("GetOffer", mock.Anything, mock.Anything).Return(offer, nil)
	svcMock.On("PutOffer", mock.Anything, expectedOffer).Return(expectedOffer, nil)

	prtMock := new(test.PrinterMock)
	prtMock.On("Print", expectedOffer).Return(nil)

	rm := new(test.RegistryMock)
	rm.On("GetCloudPartnerService").Return(svcMock, nil)
	rm.On("GetPrinter").Return(prtMock)

	cmd, err := test.QuietCommand(newPutCommand(rm))
	require.NoError(t, err)
	cmd.SetArgs([]string{"-p", offer.PublisherID, "-o", offer.ID, "-f", skuFileName})
	assert.NoError(t, cmd.Execute())
	prtMock.AssertCalled(t, "Print", expectedOffer)
}
