package main

import (
	"fmt"
	"strings"

	"github.com/klauspost/cpuid/v2"
)

// collectCPUID fills deep processor details from the CPUID instruction.
func collectCPUID(rep *Report) {
	c := cpuid.CPU

	d := CPUIDetail{}
	if c.BrandName == "" && c.PhysicalCores == 0 && c.LogicalCores == 0 {
		// Not enough info (non-x86 platform or detection failed)
		rep.CPUDetail = d
		return
	}

	d.Brand = CleanString(c.BrandName)
	d.Vendor = prettyVendor(c.VendorString, c.VendorID)
	d.FamilyModel = fmt.Sprintf("%d / %d", c.Family, c.Model)
	d.Stepping = fmt.Sprintf("%d", c.Stepping)
	d.HWThreads = c.LogicalCores
	d.PhysicalCores = c.PhysicalCores
	d.LogicalCores = c.LogicalCores
	d.ThreadsPerCore = c.ThreadsPerCore
	d.CacheLine = c.CacheLine

	if c.Cache.L1I >= 0 {
		d.L1I = fmt.Sprintf("%d KB", c.Cache.L1I)
	}
	if c.Cache.L1D >= 0 {
		d.L1D = fmt.Sprintf("%d KB", c.Cache.L1D)
	}
	if c.Cache.L2 >= 0 {
		d.L2 = fmt.Sprintf("%d KB", c.Cache.L2)
	}
	if c.Cache.L3 >= 0 {
		d.L3 = fmt.Sprintf("%d KB", c.Cache.L3)
	}

	if c.Hz > 0 {
		d.BaseClock = fmt.Sprintf("%d MHz", c.Hz/1_000_000)
	}
	if c.BoostFreq > 0 {
		d.BoostClock = fmt.Sprintf("%d MHz", c.BoostFreq/1_000_000)
	}
	if lvl := c.X64Level(); lvl > 0 {
		d.X64Level = fmt.Sprintf("x86-64-v%d", lvl)
	}

	if c.VM() {
		d.VM = "Sim (" + hypervisorName(c) + ")"
	} else {
		d.VM = "Não (físico)"
	}

	d.Features = c.FeatureSet()
	d.HasDetail = true

	rep.CPUDetail = d
}

func prettyVendor(raw string, v cpuid.Vendor) string {
	switch {
	case v == cpuid.Intel:
		return "Intel"
	case v == cpuid.AMD:
		return "AMD"
	case v == cpuid.VIA:
		return "VIA"
	case v == cpuid.Transmeta:
		return "Transmeta"
	case v == cpuid.NSC:
		return "NSC"
	}
	if raw != "" {
		return raw
	}
	return "Desconhecido"
}

func hypervisorName(c cpuid.CPUInfo) string {
	switch c.VendorID {
	case cpuid.KVM:
		return "KVM"
	case cpuid.MSVM:
		return "Microsoft Hyper-V"
	case cpuid.VMware:
		return "VMware"
	case cpuid.XenHVM:
		return "Xen"
	case cpuid.Bhyve:
		return "bhyve"
	case cpuid.Hygon:
		return "Hygon"
	}
	return "desconhecido"
}

func joinFeatures(f []string) string {
	// keep feature list compact for display
	var out []string
	for _, s := range f {
		if strings.HasPrefix(s, "(") || s == "" {
			continue
		}
		out = append(out, s)
	}
	return strings.Join(out, ", ")
}