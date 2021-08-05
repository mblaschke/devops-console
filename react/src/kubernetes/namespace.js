import React from 'react';
import $ from 'jquery';
import onClickOutside from 'react-onclickoutside'
import {CopyToClipboard} from 'react-copy-to-clipboard';

import BaseComponent from '../base';
import Spinner from '../spinner';
import Breadcrumb from '../breadcrumb';

import NamespaceDelete from './namespace.delete';
import NamespaceCreate from './namespace.create';
import NamespaceEdit from './namespace.edit';

class Namespace extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            isStartupNamespaces: true,
            config: this.buildAppConfig(),
            namespaces: [],
            confUser: {},
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
        this.ajax({
            type: "GET",
            url: '/_webapi/kubernetes/namespace'
        }).done((jqxhr) => {
            this.setState({
                namespaces: jqxhr,
                isStartupNamespaces: false
            });
        });
    }

    init() {
        this.initTeamSelection('team');
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

    handleNamespaceClick(row) {
        // close descripton if clicked somewhere else
        if (this.state.namespaceDescriptionEdit !== false && this.state.namespaceDescriptionEdit !== row.name) {
            this.handleDescriptionEditClose();
        }

        this.setState({
            selectedNamespace: row
        });
    }

    resetNamespace(namespace) {
        this.ajax({
            type: 'POST',
            url: "/_webapi/kubernetes/namespace/" + encodeURI(namespace.name) + "/reset"
        });
    }

    renderRowOwner(row) {
        let personalBadge = "";
        let teamBadge = "";
        let userBadge = "";

        if (row) {
            if (row.name && row.name.match(/^user-[^-]+-.*/i)) {
                personalBadge = <span className="badge badge-light">Personal Namespace</span>
            }

            if (row.ownerTeam && row.ownerTeam !== "") {
                teamBadge = <div><span className="badge badge-light">Team: {this.highlight(row.ownerTeam)}</span></div>
            }

            if (row.ownerUser && row.ownerUser !== "") {
                userBadge = <div><span className="badge badge-light">User: {this.highlight(row.ownerUser)}</span></div>
            }
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

    handleDescriptionEdit(row) {
        this.setState({
            namespaceDescriptionEdit: row.name,
            namespaceDescriptionEditValue: row.description
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
        this.ajax({
            type: 'PUT',
            url: "/_webapi/kubernetes/namespace/" + encodeURI(this.state.namespaceDescriptionEdit),
            data: JSON.stringify({
                description: this.state.namespaceDescriptionEditValue
            })
        }).done(() => {
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
                if (row.name.search(re) !== -1) return true;
                if (row.ownerTeam.search(re) !== -1) return true;
                if (row.ownerUser.search(re) !== -1) return true;
                if (row.description.search(re) !== -1) return true;
                if (row.createdAgo.search(re) !== -1) return true;
                if (row.status.search(re) !== -1) return true;

                if (row.settings) {
                    for(var i in row.settings) {
                        if (row.settings[i].search(re) !== -1) return true;
                    }
                }

                return false;
            });
        }

        // filter by team
        if (this.state.team !== "*") {
            ret = ret.filter((row) => {
                return row.ownerTeam === this.state.team;
            });
        }

        ret = ret.sort(function(a,b) {
            if(a.name < b.name) return -1;
            if(a.name > b.name) return 1;
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
                {this.state.config.quota.team === 0 ? 'unlimited' : this.state.config.quota.team} team /&nbsp;
                {this.state.config.quota.user === 0 ? 'unlimited' : this.state.config.quota.user} personal
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
               if (this.state.config && this.state.config.kubernetes.namespace.settings) {
                   this.state.config.kubernetes.namespace.settings.forEach((setting) => {
                       if (row.settings && row.settings[setting.name]) {
                           ret.push({
                              label: setting.label,
                              value: row.settings[setting.name]
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
            <div className="flexcontainer" onClick={this.handleDescriptionEditClose.bind(this)}>

                <Breadcrumb/>

                <div className="card mb-3 autoheight">
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
                                    <th>{this.sortBy("name", "Namespace")}</th>
                                    <th>{this.sortBy("podCount", "Pods")}</th>
                                    <th>{this.sortBy("ownerTeam", "Owner")}</th>
                                    <th>Settings</th>
                                    <th>{this.sortBy("created", "Created")}</th>
                                    <th>{this.sortBy("status", "Status")}</th>
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
                                    <tr key={row.name} className="k8s-namespace" onClick={this.handleNamespaceClick.bind(this, row)}>
                                        <td>
                                            <div className="button-copy-box">
                                                {this.highlight(row.name)}
                                                <CopyToClipboard text={row.name}>
                                                    <button className="button-copy" onClick={this.handlePreventEvent.bind(this)}><i className="far fa-copy"></i></button>
                                                </CopyToClipboard>
                                            </div>
                                            <br/>
                                            {(() => {
                                               if (this.state.namespaceDescriptionEdit === row.name) {
                                                   return <form onSubmit={this.handleDescriptionSubmit.bind(this)}>
                                                       <input type="text" className="form-control description-edit" placeholder="Description" value={this.state.namespaceDescriptionEditValue} onChange={this.handleDescriptionChange.bind(this)}/>
                                                   </form>
                                               } else {
                                                   return <small className="form-text text-muted editable description" onClick={this.handleDescriptionEdit.bind(this, row)}>{row.description ? this.highlight(row.description) : <i>no description set</i>}</small>
                                               }
                                            })()}
                                        </td>
                                        <td>
                                            <p className="text-right">{row.podCount !== null ? row.podCount : "n/a" }</p>
                                        </td>
                                        <td>
                                            {this.renderRowOwner(row)}
                                        </td>
                                        <td className="small">
                                            <div>
                                                <span className="badge badge-warning">NetworkPolicy: {row.networkPolicy || "none"}</span>
                                            </div>
                                            {namespaceSettings(row).map((setting, index) =>
                                                <div>
                                                    <span className="badge badge-light">{setting.label}: {this.highlight(setting.value)}</span>
                                                </div>
                                            )}
                                        </td>
                                        <td><div title={row.created}>{this.highlight(row.createdAgo)}</div></td>
                                        <td>
                                            {(() => {
                                                switch (row.status.toLowerCase()) {
                                                    case "terminating":
                                                        return <span className="badge badge-danger">{this.highlight(row.status)}</span>;
                                                    case "active":
                                                        return <span className="badge badge-success">{this.highlight(row.status)}</span>;
                                                    default:
                                                        return <span className="badge badge-warning">{this.highlight(row.status)}</span>;
                                                }
                                            })()}
                                            <br/>
                                            <span className={row.deleteable ? 'hidden' : 'badge badge-info'}>Not deletable</span>
                                        </td>
                                        <td className="toolbox">
                                            {(() => {
                                                switch (row.status) {
                                                case "Terminating":
                                                    return <div></div>
                                                default:
                                                    return (
                                                        <div className="btn-group" role="group">
                                                            <button id={'btnGroupDrop-' + row.name } type="button"
                                                                    className="btn btn-secondary dropdown-toggle"
                                                                    data-bs-toggle="dropdown" aria-haspopup="true"
                                                                    aria-expanded="false">
                                                                Action
                                                            </button>
                                                            <ul className="dropdown-menu" aria-labelledby={'btnGroupDrop-' + row.name }>
                                                                <li><a className="dropdown-item" onClick={self.editNamespace.bind(self, row)}>Edit</a></li>
                                                                <li><a className="dropdown-item" onClick={self.resetNamespace.bind(self, row)}>Reset Settings/RBAC</a></li>
                                                                <li><a className={row.deleteable ? 'dropdown-item' : 'hidden'} onClick={self.deleteNamespace.bind(self, row)}>Delete</a></li>
                                                            </ul>
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

                <NamespaceDelete config={this.state.config} namespace={this.state.selectedNamespaceDelete} callback={this.handleNamespaceDeletion.bind(this)} />
                <NamespaceCreate config={this.state.config} callback={this.handleNamespaceCreation.bind(this)} />
                <NamespaceEdit config={this.state.config} show={this.state.namespaceEditModalShow} namespace={this.state.selectedNamespace} callback={this.handleNamespaceEdit.bind(this)} />
            </div>
        );
    }
}

export default onClickOutside(Namespace);

