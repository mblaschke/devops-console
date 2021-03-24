import React from 'react';
import BaseComponent from '../base';
import NamespaceFormelement from "./namespace.formelement";

class NamespaceEdit extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            buttonText: "Save namespace",
            buttonState: "",
            reload: true,

            form: {}
        };
    }

    saveNamespace(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Saving...",
        });

        let jqxhr = this.ajax({
            type: 'PUT',
            url: "/_webapi/kubernetes/namespace/" + encodeURI(this.props.namespace.name),
            data: JSON.stringify(this.state.form)
        }).done((jqxhr) => {
            this.setState({
                form: false,
                reload: true,
            });

            if (this.props.callback) {
                this.props.callback()
            }
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });
    }

    cancelEdit() {
        this.setState({
            form: false,
            reload: true,
        });
    }

    componentWillReceiveProps(nextProps) {
        if (!nextProps.show) {
            // trigger hide
            this.setState({
                reload: true
            });

            return
        }

        if (!nextProps.namespace.name) {
            // invalid namespace
            return
        }

        // show modal
        if (this.state.reload) {
            // make copy
            let form = JSON.parse(JSON.stringify(nextProps.namespace));

            // set to state
            this.setState({
                form: form,
                reload: false
            });
        }
    }

    kubernetesSettingsConfig() {
        let ret = [];

        if (this.props.config.kubernetes.namespace.settings) {
            ret = this.props.config.kubernetes.namespace.settings;
        }

        return ret;
    }

    render() {
        return (
            <div>
                <form method="post">
                <div className="modal fade" id="editQuestion" tabIndex="-1" role="dialog" aria-labelledby="editQuestion" aria-hidden="true">
                    <div className="modal-dialog" role="document">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h5 className="modal-title" id="exampleModalLabel">Edit namespace</h5>
                                <button type="button" className="close" data-dismiss="modal" aria-label="Close">
                                    <span aria-hidden="true">&times;</span>
                                </button>
                            </div>
                                <div className="modal-body">
                                    <div className="form-group">
                                        <label htmlFor="inputNsApp" className="inputRg">Namespace</label>
                                        <input className="form-control" type="text" value={this.state.form.name} readOnly />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsDescription" className="inputRg">Description</label>
                                        <input type="text" name="nsDescription" id="inputNsDescription" className="form-control" placeholder="Description" value={this.getValue("form.description")} onChange={this.setValue.bind(this, "form.description")} />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsNetpol" className="inputRg">NetworkPolicy</label>
                                        <select id="inputNsNetpol" className="form-control" value={this.getValue("form.networkPolicy")} onChange={this.setValue.bind(this, "form.networkPolicy")}>
                                            {this.props.config.kubernetes.namespace.networkPolicy.map((row) =>
                                                <option key={row.name} value={row.name}>{row.description}</option>
                                            )}
                                        </select>
                                    </div>

                                    {this.kubernetesSettingsConfig().map((setting, value) =>
                                        <NamespaceFormelement setting={setting} value={this.getValue("form.settings." + setting.name)} onchange={this.setValue.bind(this, "form.settings." + setting.name)} />
                                    )}

                                <div className="modal-footer">
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.buttonState} onClick={this.saveNamespace.bind(this)}>{this.state.buttonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                </form>
            </div>
        );
    }
}

export default NamespaceEdit;

