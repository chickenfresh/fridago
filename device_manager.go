package fridago

//#include "frida-core.h"
import "C"

type DeviceManager struct {
	handle *C.FridaDeviceManager
}

func NewDeviceManager() *DeviceManager {
	return &DeviceManager{handle: C.frida_device_manager_new()}
}

func (dm *DeviceManager) Close() error {
	var gerr *C.GError

	C.frida_device_manager_close_sync(dm.handle, nil, &gerr)
	if gerr != nil {
		return NewGError(gerr)
	}

	C.g_object_unref(C.gpointer(dm.handle))
	dm.handle = nil
	return nil
}

func (dm *DeviceManager) EnumerateDevices() ([]*Device, error) {
	var gerr *C.GError

	devices := C.frida_device_manager_enumerate_devices_sync(dm.handle, nil, &gerr)
	if gerr != nil {
		return nil, NewGError(gerr)
	}

	size := int(C.frida_device_list_size(devices))
	dl := make([]*Device, size)
	for i := 0; i < size; i++ {
		fd := C.frida_device_list_get(devices, C.int(i))
		dl[i] = NewDevice(fd)
	}

	C.g_object_unref(C.gpointer(devices))
	devices = nil

	return dl, nil
}

type RemoteDeviceOptions struct {
	Certificate       string
	Origin            string
	Token             string
	KeepaliveInterval int
}

func (r RemoteDeviceOptions) SetTo(handle *C.FridaRemoteDeviceOptions) error {
	if r.Certificate != "" {
		var gerr *C.GError
		cert := C.g_tls_certificate_new_from_pem(C.CString(r.Certificate), -1, &gerr)
		if gerr != nil {
			return NewGError(gerr)
		}
		C.frida_remote_device_options_set_certificate(handle, cert)
	}
	if r.Origin != "" {
		C.frida_remote_device_options_set_origin(handle, C.CString(r.Origin))
	}
	if r.Token != "" {
		C.frida_remote_device_options_set_token(handle, C.CString(r.Token))
	}
	if r.KeepaliveInterval != 0 {
		C.frida_remote_device_options_set_keepalive_interval(handle, C.int(r.KeepaliveInterval))
	}
	return nil
}

func (dm *DeviceManager) AddRemoteDevice(address string, rds ...RemoteDeviceOptions) (*Device, error) {
	opts := C.frida_remote_device_options_new()
	defer func() {
		C.g_object_unref(C.gpointer(opts))
		opts = nil
	}()

	if len(rds) != 0 {
		err := rds[0].SetTo(opts)
		if err != nil {
			return nil, err
		}
	}

	var gerr *C.GError
	device := C.frida_device_manager_add_remote_device_sync(dm.handle, C.CString(address), opts, nil, &gerr)
	if gerr != nil {
		return nil, NewGError(gerr)
	}
	return NewDevice(device), nil
}

func (dm *DeviceManager) RemoveRemoteDevice(address string) error {
	var gerr *C.GError
	C.frida_device_manager_remove_remote_device_sync(dm.handle, C.CString(address), nil, &gerr)
	if gerr != nil {
		return NewGError(gerr)
	}
	return nil
}
