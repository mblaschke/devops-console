import React from 'react';
import BaseComponent from '../base';
import Spinner from '../spinner';
import Breadcrumb from '../breadcrumb';

class Roleassignment extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            config: this.buildAppConfig(),

            searchValue: "",

            requestRunning: false,
            form: {}
        };
    }

    resetForm() {
        let state = this.state;

        state.form = {
            resourceId: "",
            roleDefinition: "",
            ttl: "1h",
            reason: ""
        };
        let ttlList = this.ttlList();
        if (ttlList && ttlList.length >= 1) {
            state.form.ttl = ttlList[0]
        }

        this.setState(state);
    }

    init() {
        this.resetForm();
    }

    componentDidMount() {
        this.loadConfig();
        this.setInputFocus();
    }

    create(e) {
        e.preventDefault();
        e.stopPropagation();

        this.setState({
            requestRunning: true
        });

        this.ajax({
            type: 'POST',
            url: "/_webapi/azure/roleassignment",
            data: JSON.stringify(this.state.form)
        }).done(() => {
            this.resetForm();
        }).always(() => {
            this.setState({
                requestRunning: false
            });
        });
    }


    remove(e) {
        e.preventDefault();
        e.stopPropagation();

        this.setState({
            requestRunning: true
        });

        this.ajax({
            type: 'DELETE',
            url: "/_webapi/azure/roleassignment",
            data: JSON.stringify(this.state.form)
        }).done(() => {
            this.resetForm();
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });
    }

    stateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else if (this.state.form.resourceId === "" || this.state.form.roleDefinition === "" || this.state.form.ttl === "" || this.state.form.reason === "") {
            state = "disabled"
        }

        return state
    }

    handleClickOutside() {
        this.setInputFocus();
    }

    roleDefinitionList() {
        return Array.isArray(this.state.config.azure.roleAssignment.roleDefinitions) ? this.state.config.azure.roleAssignment.roleDefinitions : [];
    }

    ttlList() {
        return Array.isArray(this.state.config.azure.roleAssignment.ttl) ? this.state.config.azure.roleAssignment.ttl : [];
    }

    render() {
        if (this.state.isStartup) {
            return (
                <div>
                    <Spinner active={this.state.isStartup}/>
                </div>
            )
        }

        return (
            <div>
                <Spinner active={this.state.isStartup}/>

                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-box"></i>
                        Azure JIT access (via Azure RoleAssignment)
                    </div>
                    <div className="card-body">
                        <div className="alert alert-warning col-md-8 offset-md-2" role="alert">
                            <h4 className="alert-heading">How to JIT access</h4>
                            <p>
                                Azure RoleAssignments in general might take some minutes to be active, might not
                                automatically grant data access to Azure PaaS services by default and might require
                                additional steps:
                            </p>
                            <hr />
                            <p>
                                <dl>
                                    <dt>Azure KeyVaults</dt>
                                    <dd>
                                        Data access (secrets, certificates, ...) is not granted automatically if KeyVault
                                        is deployed in "Vault access policy" mode.
                                        In this mode access must be granted explicit inside KeyVaults "Access Policies"
                                        with Contributor permissions.
                                    </dd>
                                    <dt>Azure CosmosDB</dt>
                                    <dd>
                                        For data access Contributor permissions are required and might take several
                                        minutes to be effective.
                                    </dd>
                                </dl>
                            </p>
                        </div>
                        <form method="post">
                            <div className="form-group">
                                <label htmlFor="inputResourceId" className="inputResourceId">Azure ResourceID</label>
                                <input type="text" name="resourceId" id="inputResourceId" className="form-control"
                                       placeholder="/subscriptions/xxxxx-xxxx-xxxx-xxxx/resourceGroups/xxxxxxxx"
                                       required value={this.getValue("form.resourceId")}
                                       onChange={this.setValue.bind(this, "form.resourceId")}/>
                            </div>

                            <div className="form-group">
                                <label htmlFor="selectRoleDefinition">Role</label>
                                <select name="roleDefinition" id="selectRoleDefinition" className="form-control"
                                        value={this.getValue("form.roleDefinition")}
                                        onChange={this.setValue.bind(this, "form.roleDefinition")}>
                                    <option value="">-- select --</option>
                                    {this.roleDefinitionList().map((row, value) =>
                                        <option key={row} value={row}>{row}</option>
                                    )}
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="selectTtl">Time (ttl)</label>
                                <select name="ttl" id="selectTtl" className="form-control"
                                        value={this.getValue("form.ttl")}
                                        onChange={this.setValue.bind(this, "form.ttl")}>
                                    {this.ttlList().map((row, value) =>
                                        <option key={row} value={row}>{row}</option>
                                    )}
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputReason" className="inputReason">Reason</label>
                                <textarea className="form-control" id="inputReason" rows="3" required
                                          value={this.getValue("form.reason")}
                                          onChange={this.setValue.bind(this, "form.reason")}></textarea>
                            </div>

                            <div className="toolbox">
                                <button type="button" className="btn btn-primary" disabled={this.stateButton()}
                                        onClick={this.remove.bind(this)}>Remove RoleAssignment
                                </button>
                                <button type="button" className="btn btn-primary" disabled={this.stateButton()}
                                        onClick={this.create.bind(this)}>Create RoleAssignment
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        );
    }
}

export default Roleassignment;

