import React from 'react';
import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import Breadcrumb from "./Breadcrumb";

class K8sCluster extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            searchValue: '',
            countMasters: 0,
            countAgents: 0,
            nodes: [],
        };


        window.App.registerSearch((event) => {
            this.handleSearchChange(event);
        });
        window.App.enableSearch();

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    loadNodes() {
        let jqxhr = this.ajax({
            type: "GET",
            url: '/api/kubernetes/cluster'
        }).done((jqxhr) => {
            let countMasters = 0;
            let countAgents = 0;

            if (this.state.isStartup) {
                this.setInputFocus();
            }

            this.setState({
                nodes: jqxhr,
                isStartup: false
            });

            this.state.nodes.forEach((row) => {
                if (row.Role === "master") {
                    countMasters++;
                } else {
                    countAgents++;
                }
            });

            this.setState({
                countMasters: countMasters,
                countAgents: countAgents
            })
        });
    }

    componentDidMount() {
        this.loadNodes();
    }

    refresh() {
        this.loadNodes();
    }

    handleSearchChange(event) {
        this.setState({
            searchValue: event.target.value
        });
    }

    getNodes() {
        let ret = Array.isArray(this.state.nodes) ? this.state.nodes : [];

        if (this.state.searchValue !== "") {
            let term = this.state.searchValue.replace(/[.?*+^$[\]\\(){}|-]/g, "\\$&");
            let re = new RegExp(term, "i");

            ret = ret.filter((row) => {
                if (row.Role.search(re) !== -1) return true;
                if (row.Name.search(re) !== -1) return true;
                if (row.InternalIp.search(re) !== -1) return true;
                if (row.SpecArch.search(re) !== -1) return true;
                if (row.SpecRegion.search(re) !== -1) return true;
                if (row.SpecOS.search(re) !== -1) return true;
                if (row.SpecZone.search(re) !== -1) return true;
                if (row.SpecInstance.search(re) !== -1) return true;
                if (row.SpecMachineCPU.search(re) !== -1) return true;
                if (row.SpecMachineMemory.search(re) !== -1) return true;
                if (row.Version.search(re) !== -1) return true;
                if (row.CreatedAgo.search(re) !== -1) return true;
                if (row.Status.search(re) !== -1) return true;

                return false;
            });
        }

        ret = ret.sort(function(a,b) {
            if(a.Name < b.Name) return -1;
            if(a.Name > b.Name) return 1;
            return 0;
        });

        return ret;
    }

    render() {
        let nodes = this.getNodes();
        if (nodes) {
            return (
                <div>
                    <Breadcrumb/>

                    <div className="card mb-3">
                        <div className="card-header">
                            <i className="fas fa-server"></i>
                            Kubernetes cluster overview
                        </div>
                        <div className="card-body card-body-table spinner-area">
                            <Spinner active={this.state.isStartup}/>

                            <div className="table-responsive">
                                <table className="table table-hover table-sm spinner-area">
                                    <thead>
                                    <tr>
                                        <th>Server</th>
                                        <th>Network</th>
                                        <th>System</th>
                                        <th>Version</th>
                                        <th>Created</th>
                                        <th>Status</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {nodes.map((row) =>
                                        <tr key={row.Name} className={row.Role === 'master' ? 'k8s-master' : 'k8s-agent'}>
                                            <td>
                                                <span className={row.Role === 'master' ? 'badge badge-danger' : 'badge badge-primary'}>{this.highlight(row.Role)}</span> {this.highlight(row.Name)}<br/>
                                                <span className="badge badge-info">{this.highlight(row.SpecArch)}</span>&nbsp;
                                                <span className="badge badge-info">{this.highlight(row.SpecOS)}</span>&nbsp;
                                                <span className="badge badge-secondary">Region {this.highlight(row.SpecRegion)}</span>&nbsp;
                                                <span className="badge badge-secondary">Zone {this.highlight(row.SpecZone)}</span>
                                            </td>
                                            <td>{this.highlight(row.InternalIp)}</td>
                                            <td>
                                                <small>
                                                    {this.highlight(row.SpecInstance)}<br/>
                                                    CPU: {this.highlight(row.SpecMachineCPU)}<br/>
                                                    MEM: {this.highlight(row.SpecMachineMemory)}<br/>
                                                </small>
                                            </td>
                                            <td>{this.highlight(row.Version)}</td>
                                            <td><div title={row.Created}>{this.highlight(row.CreatedAgo)}</div></td>
                                            <td>
                                                <span
                                                    className={row.Status === 'Ready' ? 'badge badge-success' : 'badge badge-warning'}>{row.Status !== '' ? this.highlight(row.Status)  : this.highlight("unknown")}</span>
                                            </td>
                                        </tr>
                                    )}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                        <div className="card-footer small text-muted">{this.state.countMasters} masters, {this.state.countAgents} agents</div>
                    </div>
                </div>);
        }
    }
}

export default K8sCluster;

