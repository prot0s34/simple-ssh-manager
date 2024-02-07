package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"log"
	"os"
)

func handleShell(session *ssh.Session) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Printf("Failed to restore terminal: %s", err)
		}
	}()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	if err := session.RequestPty("xterm", height, width, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return fmt.Errorf("failed to request pseudo terminal: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	if err := session.Wait(); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			if exitErr.ExitStatus() == 0 {
				log.Println("SSH session finished successfully.")
			} else {
				log.Printf("SSH session exited with non-zero status: %d\n", exitErr.ExitStatus())
			}
		} else {
			return fmt.Errorf("failed to wait for session completion: %w", err)
		}
	}

	return nil
}
