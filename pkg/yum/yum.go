package yum

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bserdar/watermelon/client"
	"github.com/bserdar/watermelon/server/pb"
)

type Server struct {
	client.GRPCServer
}

func getPackages(pkgs []string, pkg string) []string {
	if len(pkgs) > 0 {
		return pkgs
	}
	if len(pkg) > 0 {
		return []string{pkg}
	}
	return []string{}
}

func (x PackageParams) getPackages() []string {
	return getPackages(x.Pkgs, x.Pkg)
}

func (s Server) Install(ctx context.Context, req *PackageParams) (*pb.Response, error) {
	packages := req.getPackages()
	if len(packages) == 0 {
		return &pb.Response{}, nil
	}
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)
	host.Logf("Yum install %+v", req)
	rsp := host.Commandf("yum install -y %s", strings.Join(packages, " "))
	scn := bufio.NewScanner(bytes.NewReader(rsp.Stdout))
	changed := true
	for scn.Scan() {
		if strings.HasPrefix(scn.Text(), "Nothing to do") {
			changed = false
			break
		}
	}
	return &pb.Response{Data: rsp.Stdout, ErrorMsg: string(rsp.Stderr), Modified: changed}, nil
}

func (s Server) Update(ctx context.Context, req *PackageParams) (*pb.Response, error) {
	packages := req.getPackages()
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)
	host.Logf("Yum update %+v", req)
	rsp := host.Commandf("yum update -y %s", strings.Join(packages, " "))
	scn := bufio.NewScanner(bytes.NewReader(rsp.Stdout))
	changed := true
	for scn.Scan() {
		if strings.HasPrefix(scn.Text(), "No packages marked") {
			changed = false
			break
		}
	}
	return &pb.Response{Data: rsp.Stdout, ErrorMsg: string(rsp.Stderr), Modified: changed}, nil
}

func (s Server) Remove(ctx context.Context, req *PackageParams) (*pb.Response, error) {
	packages := req.getPackages()
	if len(packages) == 0 {
		return &pb.Response{}, nil
	}
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)
	host.Logf("Yum remove %+v", req)
	rsp := host.Commandf("yum erase -y %s", strings.Join(packages, " "))
	scn := bufio.NewScanner(bytes.NewReader(rsp.Stdout))
	changed := true
	for scn.Scan() {
		if strings.HasPrefix(scn.Text(), "No packages marked") {
			changed = false
			break
		}
	}
	return &pb.Response{Data: rsp.Stdout, ErrorMsg: string(rsp.Stderr), Modified: changed}, nil
}

// Returns installed version, or "absent" if not installed
func getInstalled(host client.Host, pkg string) string {
	rsp := host.Commandf("yum list installed %s", pkg)
	scn := bufio.NewScanner(bytes.NewReader(rsp.Stdout))
	state := 0
	currentVersion := "absent"
	for scn.Scan() {
		text := scn.Text()
		if state == 0 {
			if strings.HasPrefix(text, "Installed Packages") {
				state = 1
			}
		} else {
			lineScn := bufio.NewScanner(strings.NewReader(text))
			lineScn.Split(bufio.ScanWords)
			if lineScn.Scan() {
				if lineScn.Scan() {
					currentVersion = lineScn.Text()
				}
			}
		}
	}
	return currentVersion
}

func (s Server) Ensure(ctx context.Context, req *EnsureParams) (*pb.Response, error) {
	log.Debugf("Yum ensure %+v", req)
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)
	log.Debugf("Running ensure on %s", host.ID)
	packages := getPackages(req.Pkgs, req.Pkg)
	if len(packages) == 0 {
		return &pb.Response{}, nil
	}

	ret := &pb.Response{Success: true}
	for _, pkg := range packages {
		if len(pkg) == 0 {
			continue
		}
		log.Debugf("Get current package")
		currentVersion := getInstalled(host, pkg)
		log.Debugf("Current version for %s is %s", pkg, currentVersion)
		if len(currentVersion) == 0 {
			return nil, fmt.Errorf("Cannot determine package status for %s", pkg)
		}
		switch req.Version {
		case "absent":
			if currentVersion != "absent" {
				w, err := s.Remove(ctx, &PackageParams{HostId: host.ID, Pkg: pkg})
				ret.Data = append(ret.Data, w.Data...)
				if w.Modified {
					ret.Modified = true
				}
				if err != nil {
					return nil, err
				}
			}

		case "installed":
			if currentVersion == "absent" {
				w, err := s.Install(ctx, &PackageParams{HostId: host.ID, Pkg: pkg})
				ret.Data = append(ret.Data, w.Data...)
				if w.Modified {
					ret.Modified = true
				}
				if err != nil {
					return nil, err
				}
			}

		case "latest":
			if currentVersion == "absent" {
				w, err := s.Install(ctx, &PackageParams{HostId: host.ID, Pkg: pkg})
				ret.Data = append(ret.Data, w.Data...)
				if w.Modified {
					ret.Modified = true
				}
				if err != nil {
					return nil, err
				}
			}
			w, err := s.Update(ctx, &PackageParams{HostId: host.ID, Pkg: pkg})
			ret.Data = append(ret.Data, w.Data...)
			if w.Modified {
				ret.Modified = true
			}
			if err != nil {
				return nil, err
			}

		default:
			if currentVersion != req.Version {
				w, err := s.Update(ctx, &PackageParams{HostId: host.ID, Pkg: pkg})
				ret.Data = append(ret.Data, w.Data...)
				if w.Modified {
					ret.Modified = true
				}
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return ret, nil
}

func (s Server) GetVer(ctx context.Context, req *GetVerParams) (*GetVerResult, error) {
	log.Debugf("Yum getver %+v", req)
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)
	log.Debugf("Running getver on %s", host.ID)

	packages := getPackages(req.Pkgs, req.Pkg)
	result := GetVerResult{}
	for _, pkg := range packages {
		currentVersion := getInstalled(host, pkg)
		if len(currentVersion) > 0 {
			result.Versions = append(result.Versions, &PkgVersion{Pkg: pkg, Version: currentVersion})
		}
	}
	return &result, nil
}
