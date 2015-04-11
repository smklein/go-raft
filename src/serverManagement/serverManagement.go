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
	Cf             config.Config
}

func StartAllServers() *ServerManager {
	sm := &ServerManager{}
	sm.serverCommands = make(map[string]*exec.Cmd)
	if !config.LoadConfig(&sm.Cf) {
		fmt.Println("[SERVER MANAGEMENT] Could not load config")
		return nil
	}

	for _, s := range sm.Cf.Servers {
		if s.Name != "debugServer" {
			fmt.Println("[SERVER MANAGEMENT] Starting server: ", s.Name)
			cmd := exec.Command("go", "run", os.Getenv("GOPATH")+"src/server/run/runner.go", s.Name)
			err := cmd.Start()
			if err != nil {
				fmt.Println("[SERVER MANAGEMENT] Could not start server")
			}
			sm.serverCommands[s.Name] = cmd
		}
	}
	return sm
}

func StartDebugServer() error {
	cmd := exec.Command("go", "run", os.Getenv("GOPATH")+"src/debugRpcServer/run/runner.go")
	return cmd.Start()
}

func (sm *ServerManager) KillAllServers() {
	for _, cmd := range sm.serverCommands {
		cmd.Process.Kill()
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
		cmd := exec.Command("go", "run", os.Getenv("GOPATH")+"src/server/run/runner.go", serverName)
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
