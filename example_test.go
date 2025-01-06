package amqp_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	amqp "github.com/hainguyen918/go-amqp"
)

func Example() {
	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqps://my-namespace.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer conn.Close()

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// send a message
	{
		// create a sender
		sender, err := session.NewSender(ctx, "/queue-name", nil)
		if err != nil {
			log.Fatal("Creating sender link:", err)
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

		// send message
		err = sender.Send(ctx, amqp.NewMessage([]byte("Hello!")), nil)
		if err != nil {
			log.Fatal("Sending message:", err)
		}

		sender.Close(ctx)
		cancel()
	}

	// continuously read messages
	{
		// create a receiver
		receiver, err := session.NewReceiver(ctx, "/queue-name", nil)
		if err != nil {
			log.Fatal("Creating receiver link:", err)
		}
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			receiver.Close(ctx)
			cancel()
		}()

		for {
			// receive next message
			msg, err := receiver.Receive(ctx, nil)
			if err != nil {
				log.Fatal("Reading message from AMQP:", err)
			}

			// accept message
			if err = receiver.AcceptMessage(context.TODO(), msg); err != nil {
				log.Fatalf("Failure accepting message: %v", err)
			}

			fmt.Printf("Message received: %s\n", msg.GetData())
		}
	}
}

func ExampleConnError() {
	// *ConnErrors are returned when the underlying connection has been closed.
	// this error is propagated to all child Session, Sender, and Receiver instances.

	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqps://my-namespace.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// create a sender
	sender, err := session.NewSender(ctx, "/queue-name", nil)
	if err != nil {
		log.Fatal("Creating sender link:", err)
	}

	// close the connection before sending the message
	conn.Close()

	// attempt to send message on a closed connection
	err = sender.Send(ctx, amqp.NewMessage([]byte("Hello!")), nil)

	var connErr *amqp.ConnError
	if !errors.As(err, &connErr) {
		log.Fatalf("unexpected error type %T", err)
	}

	// similarly, methods on session will fail in the same way
	_, err = session.NewReceiver(ctx, "/queue-name", nil)
	if !errors.As(err, &connErr) {
		log.Fatalf("unexpected error type %T", err)
	}

	// methods on the connection will also fail
	_, err = conn.NewSession(ctx, nil)
	if !errors.As(err, &connErr) {
		log.Fatalf("unexpected error type %T", err)
	}
}

func ExampleSessionError() {
	// *SessionErrors are returned when a session has been closed.
	// this error is propagated to all child Sender and Receiver instances.

	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqps://my-namespace.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer conn.Close()

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// create a sender
	sender, err := session.NewSender(ctx, "/queue-name", nil)
	if err != nil {
		log.Fatal("Creating sender link:", err)
	}

	// close the session before sending the message
	session.Close(ctx)

	// attempt to send message on a closed session
	err = sender.Send(ctx, amqp.NewMessage([]byte("Hello!")), nil)

	var sessionErr *amqp.SessionError
	if !errors.As(err, &sessionErr) {
		log.Fatalf("unexpected error type %T", err)
	}

	// similarly, methods on session will fail in the same way
	_, err = session.NewReceiver(ctx, "/queue-name", nil)
	if !errors.As(err, &sessionErr) {
		log.Fatalf("unexpected error type %T", err)
	}
}

func ExampleLinkError() {
	// *LinkError are returned by methods on Senders/Receivers after Close() has been called.
	// it can also be returned if the peer has closed the link. in this case, the *RemoteErr
	// field should contain additional information about why the peer closed the link.

	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqps://my-namespace.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer conn.Close()

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// create a sender
	sender, err := session.NewSender(ctx, "/queue-name", nil)
	if err != nil {
		log.Fatal("Creating sender link:", err)
	}

	// send message
	err = sender.Send(ctx, amqp.NewMessage([]byte("Hello!")), nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// now close the sender
	sender.Close(ctx)

	// attempt to send a message after close
	err = sender.Send(ctx, amqp.NewMessage([]byte("Hello!")), nil)

	var linkErr *amqp.LinkError
	if !errors.As(err, &linkErr) {
		log.Fatalf("unexpected error type %T", err)
	}
}

func ExampleConn_Done() {
	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqps://my-namespace.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}

	// when the channel returned by Done is closed, conn has been closed
	<-conn.Done()

	// Err indicates why the connection was closed. a nil error indicates
	// a client-side call to Close and there were no errors during shutdown.
	closedErr := conn.Err()

	// when Err returns a non-nil error, it means that either a client-side
	// call to Close encountered an error during shutdown, a fatal error was
	// encountered that caused the connection to close, or that the peer
	// closed the connection.
	if closedErr != nil {
		// the error returned by Err is always a *ConnError
		var connErr *amqp.ConnError
		errors.As(closedErr, &connErr)

		if connErr.RemoteErr != nil {
			// the peer closed the connection and provided an error explaining why.
			// note that the peer MAY send an error when closing the connection but
			// is not required to.
		} else {
			// the connection encountered a fatal error or there was
			// an error during client-side shutdown. this is for
			// diagnostics, the connection has been closed.
		}
	}
}

func ExampleSender_SendWithReceipt() {
	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqps://my-namespace.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer conn.Close()

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// create a sender
	sender, err := session.NewSender(ctx, "/queue-name", nil)
	if err != nil {
		log.Fatal("Creating sender link:", err)
	}

	// send message
	receipt, err := sender.SendWithReceipt(ctx, amqp.NewMessage([]byte("Hello!")), nil)
	if err != nil {
		log.Fatal("Sending message:", err)
	}

	// wait for confirmation of settlement
	state, err := receipt.Wait(ctx)
	if err != nil {
		log.Fatal("Wait on receipt:", err)
	}

	// determine how the peer settled the message
	switch stateType := state.(type) {
	case *amqp.StateAccepted:
		// message was accepted, no further action is required
	case *amqp.StateModified:
		// message must be modified and resent before it can be processed.
		// the values in stateType provide further context.
	case *amqp.StateReceived:
		// see the fields in [StateReceived] for information on
		// how to handle this delivery state.
	case *amqp.StateRejected:
		// the peer rejected the message
		if stateType.Error != nil {
			// the error will provide information about why the
			// message was rejected. note that the peer isn't required
			// to provide an error.
		}
	case *amqp.StateReleased:
		// message was not and will not be acted upon
	}
}
