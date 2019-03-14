package register

import (
	"github.com/iofog/iofog-kubelet/providers"
	"github.com/iofog/iofog-kubelet/providers/web"
)

func init() {
	register("web", initWeb)
}

func initWeb(cfg InitConfig) (providers.Provider, error) {
	return web.NewBrokerProvider(
		cfg.DaemonPort,
		cfg.NodeName,
		cfg.OperatingSystem,
		cfg.ControllerToken,
		cfg.ControllerUrl,
		cfg.NodeId)
}
