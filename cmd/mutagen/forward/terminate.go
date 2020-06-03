package forward

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/mutagen-io/mutagen/cmd"
	"github.com/mutagen-io/mutagen/cmd/mutagen/daemon"

	"github.com/mutagen-io/mutagen/pkg/grpcutil"
	"github.com/mutagen-io/mutagen/pkg/selection"
	forwardingsvc "github.com/mutagen-io/mutagen/pkg/service/forwarding"
)

// TerminateWithLabelSelector is an orchestration convenience method invokes the
// terminate command using the specified label selector.
func TerminateWithLabelSelector(labelSelector string) error {
	terminateConfiguration.labelSelector = labelSelector
	return terminateMain(nil, nil)
}

// TerminateWithSelection is an orchestration convenience method invokes
// terminate using the provided service client and session specification.
func TerminateWithSelection(
	client forwardingsvc.ForwardingClient,
	selection *selection.Selection,
) error {
	// Invoke the terminate method and defer closure of the RPC stream.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.Terminate(ctx)
	if err != nil {
		return errors.Wrap(grpcutil.PeelAwayRPCErrorLayer(err), "unable to invoke terminate")
	}

	// Send the initial request.
	request := &forwardingsvc.TerminateRequest{
		Selection: selection,
	}
	if err := stream.Send(request); err != nil {
		return errors.Wrap(grpcutil.PeelAwayRPCErrorLayer(err), "unable to send terminate request")
	}

	// Create a status line printer.
	statusLinePrinter := &cmd.StatusLinePrinter{}

	// Receive and process responses until we're done.
	for {
		if response, err := stream.Recv(); err != nil {
			statusLinePrinter.BreakIfNonEmpty()
			return errors.Wrap(grpcutil.PeelAwayRPCErrorLayer(err), "terminate failed")
		} else if err = response.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid terminate response received")
		} else if response.Message == "" {
			statusLinePrinter.Clear()
			return nil
		} else if response.Message != "" {
			statusLinePrinter.Print(response.Message)
			if err := stream.Send(&forwardingsvc.TerminateRequest{}); err != nil {
				statusLinePrinter.BreakIfNonEmpty()
				return errors.Wrap(grpcutil.PeelAwayRPCErrorLayer(err), "unable to send message response")
			}
		}
	}
}

// terminateMain is the entry point for the terminate command.
func terminateMain(_ *cobra.Command, arguments []string) error {
	// Create session selection specification.
	selection := &selection.Selection{
		All:            terminateConfiguration.all,
		Specifications: arguments,
		LabelSelector:  terminateConfiguration.labelSelector,
	}
	if err := selection.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid session selection specification")
	}

	// Connect to the daemon and defer closure of the connection.
	daemonConnection, err := daemon.Connect(true, true)
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	// Create a session service client.
	sessionService := forwardingsvc.NewForwardingClient(daemonConnection)

	// Perform the terminate operation.
	return TerminateWithSelection(sessionService, selection)
}

// terminateCommand is the terminate command.
var terminateCommand = &cobra.Command{
	Use:          "terminate [<session>...]",
	Short:        "Permanently terminate a forwarding session",
	RunE:         terminateMain,
	SilenceUsage: true,
}

// terminateConfiguration stores configuration for the terminate command.
var terminateConfiguration struct {
	// help indicates whether or not to show help information and exit.
	help bool
	// all indicates whether or not all sessions should be terminated.
	all bool
	// labelSelector encodes a label selector to be used in identifying which
	// sessions should be paused.
	labelSelector string
}

func init() {
	// Grab a handle for the command line flags.
	flags := terminateCommand.Flags()

	// Disable alphabetical sorting of flags in help output.
	flags.SortFlags = false

	// Manually add a help flag to override the default message. Cobra will
	// still implement its logic automatically.
	flags.BoolVarP(&terminateConfiguration.help, "help", "h", false, "Show help information")

	// Wire up terminate flags.
	flags.BoolVarP(&terminateConfiguration.all, "all", "a", false, "Terminate all sessions")
	flags.StringVar(&terminateConfiguration.labelSelector, "label-selector", "", "Terminate sessions matching the specified label selector")
}
