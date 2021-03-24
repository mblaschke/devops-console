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
            buttonText: "Create Azure RoleAssignment",

            requestRunning: false,
            form: {
                resourceId: "",
                roleDefinition: "",
                reason: ""
            }
        };
    }

    init() {
        this.componentWillMount();
    }

    componentWillMount() {
    }

    componentDidMount() {
        this.loadConfig();
        this.setInputFocus();
    }

    create(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            requestRunning: true,
            buttonText: "Creating..."
        });

        this.ajax({
            type: 'POST',
            url: "/_webapi/azure/roleassignment",
            data: JSON.stringify(this.state.form)
        }).done(() => {
            let state = this.state;
            state.form = {
                resourceId: "",
                roleDefinition: "",
                reason: ""
            };
            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });
    }

    stateCreateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else if (this.state.form.resourceId === "" || this.state.form.roleDefinition === "" || this.state.form.reason === "") {
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
                        Create Azure RoleAssignment
                    </div>
                    <div className="card-body">
                        <form method="post">
                            <div className="form-group">
                                <label htmlFor="inputResourceId" className="inputResourceId">Azure ResourceID</label>
                                <input type="text" name="resourceId" id="inputResourceId" className="form-control" placeholder="/subscription/xxxxx-xxxx-xxxx-xxxx/resourceGroup/xxxxxxxx/..." required value={this.getValue("form.resourceId")} onChange={this.setValue.bind(this, "form.resourceId")} />
                            </div>

                            <div className="form-group">
                                <label htmlFor="selectRoleDefinition">Role</label>
                                <select name="roleDefinition" id="selectRoleDefinition" className="form-control" value={this.getValue("form.roleDefinition")} onChange={this.setValue.bind(this, "form.roleDefinition")}>
                                    <option value="">-- select --</option>
                                    {this.roleDefinitionList().map((row, value) =>
                                        <option key={row} value={row}>{row}</option>
                                    )}
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputReason" className="inputReason">Reason</label>
                                <textarea className="form-control" id="inputReason" rows="3" required value={this.getValue("form.reason")} onChange={this.setValue.bind(this, "form.reason")}></textarea>
                            </div>

                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateCreateButton()} onClick={this.create.bind(this)}>{this.state.buttonText}</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        );
    }
}

export default Roleassignment;

