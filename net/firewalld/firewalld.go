package firewalld

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/bserdar/watermelon/client"
	"github.com/bserdar/watermelon/server/pb"
)

type Server struct {
	client.GRPCServer
}

func (adr RichRuleAddress) ToString() string {
	c := fmt.Sprintf(`address="%s"`, adr.Address)
	if adr.Invert {
		return "not " + c
	}
	return c
}

func (p RichRulePort) ToString() string {
	return fmt.Sprintf(`port="%s" protocol="%s"`, p.Port, p.Protocol)
}

func (r RichRule) ToString() string {
	out := &bytes.Buffer{}
	out.WriteString("rule ")
	if len(r.Family) > 0 {
		fmt.Fprintf(out, `family="%s" `, r.Family)
	}
	if r.Source != nil {
		fmt.Fprintf(out, "source %s ", r.Source.ToString())
	}
	if r.Dest != nil {
		fmt.Fprintf(out, "destination %s ", r.Dest.ToString())
	}
	if len(r.ServiceName) > 0 {
		fmt.Fprintf(out, `service name="%s" `, r.ServiceName)
	}
	if r.Port != nil {
		fmt.Fprintf(out, "port %s ", r.Port.ToString())
	}
	if r.SourcePort != nil {
		fmt.Fprintf(out, "source-port %s ", r.SourcePort.ToString())
	}
	if r.ForwardPort != nil {
		fmt.Fprintf(out, "forward-port %s", r.ForwardPort.ToString())
	}
	if len(r.Protocol) > 0 {
		fmt.Fprintf(out, `protocol value="%s" `, r.Protocol)
	}
	if len(r.Action) > 0 {
		fmt.Fprintf(out, "%s ", r.Action)
	}

	return strings.TrimSpace(out.String())
}

func (s Server) AddRule(ctx context.Context, req *AddRuleRequest) (*pb.Response, error) {
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)

	if r := req.GetRich(); r != nil {
		zone := ""
		if len(req.Zone) > 0 {
			zone = fmt.Sprintf("--zone=%s", req.Zone)
		}
		perm := ""
		if req.Permanent {
			perm = " --permanent"
		}

		rsp := host.Commandf(`firewall-cmd %s --query-rich-rule '%s' %s`, zone, r.ToString(), perm)
		if strings.Index(rsp.Out(), "yes") != -1 {
			return &pb.Response{Success: true}, nil
		}
		rsp = host.Commandf(`firewall-cmd %s --add-rich-rule '%s' %s`, zone, r.ToString(), perm)
		if len(rsp.Stderr) > 0 {
			return &pb.Response{ErrorMsg: string(rsp.Stderr)}, nil
		}
		return &pb.Response{Success: true, Modified: true}, nil
	}
	return &pb.Response{Success: true}, nil
}

func (s Server) Reload(ctx context.Context, req *ReloadRequest) (*pb.Response, error) {
	session := s.SessionFromContext(ctx)
	host := session.Host(req.HostId)
	host.Command("firewall-cmd --reload")
	return &pb.Response{}, nil
}
