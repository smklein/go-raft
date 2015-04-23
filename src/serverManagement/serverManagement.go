package serverManagement

import (
	"config"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type ServerManager struct {
	serverCommands map[string]*exec.Cmd
	debugServer    *exec.Cmd
	Cf             config.Config
}

func (sm *ServerManager) StartAllServers() error {
	sm.serverCommands = make(map[string]*exec.Cmd)
	if !config.LoadConfig(&sm.Cf) {
		fmt.Println("[SERVER MANAGEMENT] Could not load config")
		return errors.New("Could not load config")
	}

	for _, s := range sm.Cf.Servers {
		if s.Name != "debugServer" {
			fmt.Println("[SERVER MANAGEMENT] Starting server: ", s.Name)
			cmd := exec.Command(os.Getenv("GOPATH")+"bin/runner", s.Name)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Start()
			if err != nil {
				fmt.Println("[SERVER MANAGEMENT] Could not start server")
				return err
			} else {
				fmt.Println("[SERVER MANAGEMENT] Server Started")
			}
			sm.serverCommands[s.Name] = cmd
			go cmd.Wait()
		}
	}
	return nil
}

func (sm *ServerManager) StartDebugServer() error {
	sm.debugServer = exec.Command(os.Getenv("GOPATH") + "bin/debug_runner")
	sm.debugServer.Stdout = os.Stdout
	sm.debugServer.Stderr = os.Stderr
	err := sm.debugServer.Start()
	go sm.debugServer.Wait()
	return err
}

func (sm *ServerManager) KillAllServers() {
	for _, cmd := range sm.serverCommands {
		err := cmd.Process.Kill()
		if err != nil {
			panic(err)
		}
	}
	err := sm.debugServer.Process.Kill()
	if err != nil {
		panic(err)
	}
}

func (sm *ServerManager) KillServer(serverName string) error {
	if cmd, ok := sm.serverCommands[serverName]; ok {
		if err := cmd.Process.Kill(); err != nil {
			return errors.New("Could not kill process")
		}
		return nil
	} else {
		return errors.New("Server name not recognized")
	}
}

func (sm *ServerManager) RestartAllServers() error {
	for _, s := range sm.Cf.Servers {
		if err := sm.RestartServer(s.Name); err != nil {
			return err
		}
	}
	return nil
}

func (sm *ServerManager) RestartServer(serverName string) error {
	if _, ok := sm.serverCommands[serverName]; ok {
		cmd := exec.Command(os.Getenv("GOPATH")+"bin/runner", serverName)
		if err := cmd.Start(); err != nil {
			return errors.New("Could not restart server")
		} else {
			sm.serverCommands[serverName] = cmd
			return nil
		}
	} else {
		return errors.New("Server name not recognized")
	}
}
