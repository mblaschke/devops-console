import BaseComponent from '../base';
import Spinner from '../spinner';
import Breadcrumb from '../breadcrumb';

class SupportPagerduty extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            config: this.buildAppConfig(),

            form: {
                team: "",
                endpoint: "",
            }
        };
    }

    init() {
        this.componentWillMount();
    }

    componentWillMount() {
        this.initTeamSelection('form.team');
    }

    componentDidMount() {
        this.loadConfig();
        this.setInputFocus();
    }

    stateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else if (
            this.state.form.team === ""
            || this.state.form.endpoint === ""
            || this.state.form.resourceType === ""
            || this.state.form.resourceGroup === ""
            || this.state.form.location === ""
            || this.state.form.resource === ""
            || this.state.form.message === ""
        ) {
            state = "disabled"
        }

        return state
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
            url: "/_webapi/support/pagerduty",
            data: JSON.stringify(this.state.form)
        }).done((jqxhr) => {
            let state = this.state;
            state.form = {
                location: "",
                endpoint: "",
                resourceType: "",
                resourceGroup: "",
                resource: "",
                message: "",
                contact: ""
            };
            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });
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
                        Create PagerDuty emergency ticket
                    </div>
                    <div className="card-body">
                        <div className="alert alert-warning" role="alert">
                            You are about to create an emergency support ticket!
                        </div>

                        <form method="post">
                            <div className="row">
                                <div className="col col-5">
                                    <div className="form-group">
                                        <label htmlFor="inputTeam">Request assistance for team</label>
                                        <select name="inputTeam" id="inputTeam"
                                                className="form-control" value={this.getValue("form.team")}
                                                onChange={this.setTeam.bind(this, "form.team")}>
                                            <option key="" value="">- please select -</option>
                                            {this.state.config.teams.map((row, value) =>
                                                <option key={row.Id} value={row.name}>{row.name}</option>
                                            )}
                                        </select>
                                    </div>
                                </div>
                                <div className="col col-2 d-flex align-items-center justify-content-center">
                                    <i className="fa-solid fa-right-long size-2rem"></i>
                                </div>
                                <div className="col col-5">
                                    <div className="form-group">
                                        <label htmlFor="inputEndpoint">PagerDuty service</label>
                                        <select name="inputEndpoint" id="inputEndpoint"
                                                className="form-control" value={this.getValue("form.endpoint")}
                                                onChange={this.setValue.bind(this, "form.endpoint")}>
                                            <option key="" value="">- please select -</option>
                                            {this.state.config.support.pagerduty.endpoints.map((row, value) =>
                                                <option key={row.id} value={row.id}>{row.name}</option>
                                            )}
                                        </select>
                                    </div>
                                </div>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputResourceType" className="inputResourceType">Resource type</label>
                                <select name="inputResourceType" id="inputResourceType"
                                        className="form-control"
                                        value={this.getValue("form.resourceType")}
                                        onChange={this.setValue.bind(this, "form.resourceType")}>
                                    <option key="" value="">- please select -</option>
                                    <option key="Azure" value="Azure">Azure</option>
                                    <option key="Kubernetes" value="Kubernetes">Kubernetes</option>
                                    <option key="Other" value="Other">Other</option>
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputLocation" className="inputLocation">Location</label>
                                <input type="text" name="location" id="inputLocation" className="form-control" required
                                       value={this.getValue("form.location")}
                                       onChange={this.setValue.bind(this, "form.location")}/>
                                <small className="form-text text-muted">Azure subscription / Kubernetes cluster</small>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputResource"
                                       className="inputResourceGroup">ResourceGroup/Namespace</label>
                                <input type="text" name="resource" id="inputResourceGroup" className="form-control"
                                       required value={this.getValue("form.resourceGroup")}
                                       onChange={this.setValue.bind(this, "form.resourceGroup")}/>
                                <small className="form-text text-muted">Resource Group / Kubernetes namespace</small>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputResource" className="inputResource">Resource/component</label>
                                <input type="text" name="resource" id="inputResource" className="form-control" required
                                       value={this.getValue("form.resource")}
                                       onChange={this.setValue.bind(this, "form.resource")}/>
                                <small className="form-text text-muted">Resource ID / name</small>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputMessage" className="inputMessage">Message</label>
                                <textarea className="form-control" id="inputMessage" rows="3" required
                                          value={this.getValue("form.message")}
                                          onChange={this.setValue.bind(this, "form.message")}></textarea>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputContact" className="inputContact">Contact</label>
                                <textarea className="form-control" id="inputContact" rows="3" required
                                          value={this.getValue("form.contact")}
                                          onChange={this.setValue.bind(this, "form.contact")}></textarea>
                            </div>

                            <div className="toolbox">
                                <button type="button" className="btn btn-primary" disabled={this.stateButton()}
                                        onClick={this.create.bind(this)}>Create emergency ticket
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        );
    }
}

export default SupportPagerduty;
