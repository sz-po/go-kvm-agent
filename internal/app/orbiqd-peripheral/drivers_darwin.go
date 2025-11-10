package orbiqd_peripheral

import (
	"github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-peripheral/peripheral/ffmpeg"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/driver"
	driverSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/driver"
)

func createDriverRepository() (driverSDK.DriverRepository, error) {
	return driver.NewLocalRepository(
		driver.WithDriver(ffmpeg.DisplaySinkDriver),
		driver.WithDriver(ffmpeg.DisplaySourceDriver),
	)
}
