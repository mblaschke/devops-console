import React from 'react';
import $ from 'jquery';
import onClickOutside from 'react-onclickoutside'
import {CopyToClipboard} from 'react-copy-to-clipboard';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import K8sNamespaceModalDelete from './K8sNamespaceModalDelete';
import K8sNamespaceModalCreate from './K8sNamespaceModalCreate';
import K8sNamespaceModalEdit from './K8sNamespaceModalEdit';
import Breadcrumb from './Breadcrumb';

class K8sNamespace extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            isStartupNamespaces: true,
            namespaces: [],
            confUser: {},
            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                NamespaceEnvironments: [],
                Quota: {},
                Kubernetes: {
                    Namespace: {
                        Settings: [],
                        NetworkPolicy: []
                    }
                }
            },
            team: "*",
            namespaceDescriptionEdit: false,
            namespaceDescriptionEditValue: "",
            namespaceEditModalShow: false,
            selectedNamespace: [],
            selectedNamespaceDelete: [],
            namespacePreview: "",
            searchValue: "",
        };

        window.App.registerSearch((event) => {
            this.handleSearchChange(event);
        });

        window.App.enableSearch();

        $(document).on('show.bs.modal', ".modal", () => {
            this.disableRefresh();
        });
        $(document).on('hide.bs.modal', ".modal", () => {
            this.setState({
                namespaceEditModalShow: false,
                isStartupNamespaces: true
            });
            this.refresh();
        });
    }

    loadNamespaces() {
        let jqxhr = this.ajax({
            type: "GET",
            url: '/_webapi/kubernetes/namespace'
        }).done((jqxhr) => {
            this.setState({
                namespaces: jqxhr,
                isStartupNamespaces: false
            });
        });
    }

    loadConfig() {
        let jqxhr = this.ajax({
            type: "GET",
            url: '/_webapi/app/config'
        }).done((jqxhr) => {
            if (jqxhr) {
                if (this.state.isStartup) {
                    this.setInputFocus();
                }

                if (!jqxhr.Teams) {
                    jqxhr.Teams = [];
                }

                if (!jqxhr.NamespaceEnvironments) {
                    jqxhr.NamespaceEnvironments = [];
                }

                this.setState({
                    isStartup: false,
                    config: jqxhr
                });

                // trigger init
                setTimeout(() => {
                    this.init();
                });
            }
        });
    }

    init() {
        this.initTeam();
        this.refresh();
    }

    componentDidMount() {
        this.loadConfig();
    }

    refresh() {
        this.loadNamespaces();

        try {
            clearTimeout(this.refreshHandler);
        } catch(e) {}

        this.refreshHandler = setTimeout(() =>{
            this.refresh();
        }, 10000);
    }

    disableRefresh() {
        try {
            clearTimeout(this.refreshHandler);
        } catch(e) {}
    }

    handleClickOutside() {
        this.handleDescriptionEditClose();
    }

    deleteNamespace(row) {
        this.setState({
            selectedNamespaceDelete: row
        });

        setTimeout(() => {
            $("#deleteQuestion").modal('show');
            setTimeout(() => {
                $("#deleteQuestion").find(":input:text:visible:enabled").first().focus();
            },500);
        }, 200);
    }

    createNamespace() {
        setTimeout(() => {
            $("#createQuestion").modal('show');
            setTimeout(() => {
                $("#createQuestion").find(":input:text:visible:enabled").first().focus();
            },500);
        }, 200);
    }

    editNamespace(namespace) {
        this.setState({
            namespaceEditModalShow: true,
            selectedNamespace: namespace
        });

        setTimeout(() => {
            $("#editQuestion").modal('show');
        }, 200)
    }

    handleNamespaceClick(row, event) {
        // close descripton if clicked somewhere else
        if (this.state.namespaceDescriptionEdit !== false && this.state.namespaceDescriptionEdit !== row.Name) {
            this.handleDescriptionEditClose();
        }

        this.setState({
            selectedNamespace: row
        });
    }

    resetNamespace(namespace) {
        let jqxhr = this.ajax({
            type: 'POST',
            url: "/_webapi/kubernetes/namespace/" + encodeURI(namespace.Name) + "/reset"
        });
    }

    renderRowOwner(row) {
        let personalBadge = "";
        let teamBadge = "";
        let userBadge = "";

        if (row.Name.match(/^user-[^-]+-.*/i)) {
            personalBadge = <span className="badge badge-light">Personal Namespace</span>
        }

        if (row.OwnerTeam !== "") {
            teamBadge = <div><span className="badge badge-light">Team: {this.highlight(row.OwnerTeam)}</span></div>
        }

        if (row.OwnerUser !== "") {
            userBadge = <div><span className="badge badge-light">User: {this.highlight(row.OwnerUser)}</span></div>
        }

        return <span className="badge-list">{personalBadge}{teamBadge}{userBadge}</span>
    }

    handleNamespaceDeletion() {
        $("#deleteQuestion").modal('hide');
        this.setState({
            selectedNamespace: [],
        });
    }

    handleNamespaceCreation() {
        $("#createQuestion").modal('hide');
        this.setState({
            selectedNamespace: [],
        });
    }

    handleNamespaceEdit() {
        this.setState({
            selectedNamespace: [],
            namespaceEditModalShow: false
        });

        setTimeout(() => {
            $("#editQuestion").modal('hide');
        }, 200);
    }

    handleDescriptionEditClose() {
        this.setState({
            namespaceDescriptionEdit: false,
            namespaceDescriptionEditValue: ""
        });
    }

    handleDescriptionEdit(row, event) {
        this.setState({
            namespaceDescriptionEdit: row.Name,
            namespaceDescriptionEditValue: row.Description
        });

        setTimeout(() => {
           $(".description-edit:input").focus();
        },250);
    }

    handleDescriptionChange(event) {
        this.setState({
            namespaceDescriptionEditValue: event.target.value
        });
    }

    handleDescriptionSubmit(event) {
        let jqxhr = this.ajax({
            type: 'PUT',
            url: "/_webapi/kubernetes/namespace/" + encodeURI(this.state.namespaceDescriptionEdit),
            data: JSON.stringify({
                description: this.state.namespaceDescriptionEditValue
            })
        }).done((jqxhr) => {
            this.setState({
                namespaceDescriptionEdit: false
            });
            this.refresh();
        });

        event.preventDefault();
        event.stopPropagation();
        return false;
    }

    getNamespaces() {
        let ret = Array.isArray(this.state.namespaces) ? this.state.namespaces : [];

        if (this.state.searchValue !== "") {
            let term = this.state.searchValue.replace(/[.?*+^$[\]\\(){}|-]/g, "\\$&");
            let re = new RegExp(term, "i");

            ret = ret.filter((row) => {
                if (row.Name.search(re) !== -1) return true;
                if (row.OwnerTeam.search(re) !== -1) return true;
                if (row.OwnerUser.search(re) !== -1) return true;
                if (row.Description.search(re) !== -1) return true;
                if (row.CreatedAgo.search(re) !== -1) return true;
                if (row.Status.search(re) !== -1) return true;

                if (row.Settings) {
                    for(var i in row.Settings) {
                        if (row.Settings[i].search(re) !== -1) return true;
                    }
                }

                return false;
            });
        }

        // filter by team
        if (this.state.team !== "*") {
            ret = ret.filter((row) => {
                if (row.OwnerTeam === this.state.team) return true;
                return false;
            });
        }

        ret = ret.sort(function(a,b) {
            if(a.Name < b.Name) return -1;
            if(a.Name > b.Name) return 1;
            return 0;
        });
        ret = this.sortDataset(ret);

        return ret;
    }


    buildFooter(namespaces) {
        let NamespaceCountTotal = this.state.namespaces.length;
        let NamespaceCountShown = namespaces.length;

        return (
            <span>
                Namespaces: {NamespaceCountShown} of {NamespaceCountTotal},&nbsp;
                Quota:&nbsp;
                {this.state.config.Quota.team === 0 ? 'unlimited' : this.state.config.Quota.team} team /&nbsp;
                {this.state.config.Quota.user === 0 ? 'unlimited' : this.state.config.Quota.user} personal
            </span>
        )
    }

    render() {
        if (this.state.isStartup) {
            return (
                <div>
                    <Breadcrumb/>
                    <Spinner active={this.state.isStartup}/>
                </div>
            )
        }

        let namespaceSettings = (row) => {
           let ret = [];
           try {
               if (this.state.config && this.state.config.Kubernetes.Namespace.Settings) {
                   this.state.config.Kubernetes.Namespace.Settings.map((setting) => {
                       if (row.Settings && row.Settings[setting.Name]) {
                           ret.push({
                              Label: setting.Label,
                              Value: row.Settings[setting.Name]
                           });
                       }
                   });
               }
           } catch (e) {}
           return ret;
        };

        let self = this;
        let namespaces = this.getNamespaces();
        return (
            <div onClick={this.handleDescriptionEditClose.bind(this)}>

                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-object-group"></i>
                        Kubernetes namespaces
                        <div className="toolbox">
                                <div className="form-group row">
                                <div className="col-sm-8"></div>
                                <div className="col-sm-4">
                                    {this.renderTeamSelector()}
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="card-body card-body-table spinner-area">
                        <Spinner active={this.state.isStartupNamespaces}/>

                        <div className="table-responsive">
                            <table className="table table-hover table-sm">
                                <colgroup>
                                    <col width="*" />
                                    <col width="50rem" />
                                    <col width="200rem" />
                                    <col width="200rem" />
                                    <col width="200rem" />
                                    <col width="100rem" />
                                    <col width="80rem" />
                                </colgroup>
                                <thead>
                                <tr>
                                    <th>{this.sortBy("Name", "Namespace")}</th>
                                    <th>{this.sortBy("PodCount", "Pods")}</th>
                                    <th>{this.sortBy("OwnerTeam", "Owner")}</th>
                                    <th>Settings</th>
                                    <th>{this.sortBy("Created", "Created")}</th>
                                    <th>{this.sortBy("Status", "Status")}</th>
                                    <th className="toolbox">
                                        <button type="button" className="btn btn-secondary" onClick={this.createNamespace.bind(this)}>
                                            <i className="fas fa-plus"></i>
                                        </button>
                                    </th>
                                </tr>
                                </thead>
                                <tbody>
                                {namespaces.length === 0 &&
                                <tr>
                                    <td colspan="7" className="not-found">No namespaces found.</td>
                                </tr>
                                }
                                {namespaces.map((row) =>
                                    <tr key={row.Name} className="k8s-namespace" onClick={this.handleNamespaceClick.bind(this, row)}>
                                        <td>
                                            <div class="button-copy-box">
                                                {this.highlight(row.Name)}
                                                <CopyToClipboard text={row.Name}>
                                                    <button className="button-copy" onClick={this.handlePreventEvent.bind(this)}><i className="far fa-copy"></i></button>
                                                </CopyToClipboard>
                                            </div>
                                            <br/>
                                            {(() => {
                                               if (this.state.namespaceDescriptionEdit === row.Name) {
                                                   return <form onSubmit={this.handleDescriptionSubmit.bind(this)}>
                                                       <input type="text" className="form-control description-edit" placeholder="Description" value={this.state.namespaceDescriptionEditValue} onChange={this.handleDescriptionChange.bind(this)}/>
                                                   </form>
                                               } else {
                                                   return <small className="form-text text-muted editable description" onClick={this.handleDescriptionEdit.bind(this, row)}>{row.Description ? this.highlight(row.Description) : <i>no description set</i>}</small>
                                               }
                                            })()}
                                        </td>
                                        <td>
                                            <p className="text-right">{row.PodCount !== null ? row.PodCount : "n/a" }</p>
                                        </td>
                                        <td>
                                            {this.renderRowOwner(row)}
                                        </td>
                                        <td class="small">
                                            <div>
                                                <span className="badge badge-warning">NetworkPolicy: {row.NetworkPolicy || "none"}</span>
                                            </div>
                                            {namespaceSettings(row).map((setting, index) =>
                                                <div>
                                                    <span className="badge badge-light">{setting.Label}: {this.highlight(setting.Value)}</span>
                                                </div>
                                            )}
                                        </td>
                                        <td><div title={row.Created}>{this.highlight(row.CreatedAgo)}</div></td>
                                        <td>
                                            {(() => {
                                                switch (row.Status.toLowerCase()) {
                                                    case "terminating":
                                                        return <span className="badge badge-danger">{this.highlight(row.Status)}</span>;
                                                    case "active":
                                                        return <span className="badge badge-success">{this.highlight(row.Status)}</span>;
                                                    default:
                                                        return <span className="badge badge-warning">{this.highlight(row.Status)}</span>;
                                                }
                                            })()}
                                            <br/>
                                            <span className={row.Deleteable ? 'hidden' : 'badge badge-info'}>Not deletable</span>
                                        </td>
                                        <td className="toolbox">
                                            {(() => {
                                                switch (row.Status) {
                                                case "Terminating":
                                                    return <div></div>
                                                default:
                                                    return (
                                                        <div className="btn-group" role="group">
                                                            <button id={'btnGroupDrop-' + row.Name } type="button"
                                                                    className="btn btn-secondary dropdown-toggle"
                                                                    data-toggle="dropdown" aria-haspopup="true"
                                                                    aria-expanded="false">
                                                                Action
                                                            </button>
                                                            <div className="dropdown-menu" aria-labelledby={'btnGroupDrop-' + row.Name }>
                                                                <a className="dropdown-item" onClick={self.editNamespace.bind(self, row)}>Edit</a>
                                                                <a className="dropdown-item" onClick={self.resetNamespace.bind(self, row)}>Reset Settings/RBAC</a>
                                                                <a className={row.Deleteable ? 'dropdown-item' : 'hidden'} onClick={self.deleteNamespace.bind(self, row)}>Delete</a>
                                                            </div>
                                                        </div>
                                                    );
                                                }
                                            })()}
                                        </td>
                                    </tr>
                                )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div className="card-footer small text-muted">{this.buildFooter(namespaces)}</div>
                </div>

                <K8sNamespaceModalDelete config={this.state.config} namespace={this.state.selectedNamespaceDelete} callback={this.handleNamespaceDeletion.bind(this)} />
                <K8sNamespaceModalCreate config={this.state.config} callback={this.handleNamespaceCreation.bind(this)} />
                <K8sNamespaceModalEdit config={this.state.config} show={this.state.namespaceEditModalShow} namespace={this.state.selectedNamespace} callback={this.handleNamespaceEdit.bind(this)} />
            </div>
        );
    }
}

export default onClickOutside(K8sNamespace);

