import React from 'react';
import BaseComponent from './BaseComponent';
import K8sNamespaceFormElement from "./K8sNamespaceFormElement";

class K8sNamespaceModalEdit extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            buttonText: "Save namespace",
            buttonState: "",
            reload: true,

            namespace: {}
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
            url: "/api/kubernetes/namespace/" + encodeURI(this.props.namespace.Name),
            data: JSON.stringify(this.state.namespace)
        }).done((jqxhr) => {
            this.setState({
                namespace: false,
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
            namespace: false,
            reload: true,
        });
    }

    handleNamespaceInputChange(name, event) {
        var state = this.state;
        state.namespace[name] = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);
    }

    handleNamespaceSettingInputChange(name, event) {
        var state = this.state;

        if (!state.namespace.Settings) {
            state.namespace.Settings = {}
        }

        state.namespace.Settings[name] = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);
    }

    getNamespaceItem(name) {
        var ret = "";

        if (this.state.namespace && this.state.namespace[name]) {
            ret = this.state.namespace[name];
        }

        return ret;
    }

    getNamespaceSettingItem(name) {
        var ret = "";

        if (this.state.namespace.Settings && this.state.namespace.Settings[name]) {
            ret = this.state.namespace.Settings[name];
        }

        return ret;
    }

    handleNsDescriptionChange(event) {
        let state = this.state;
        state.namespace.description = event.target.value;
        this.setState(state);
    }

    componentWillReceiveProps(nextProps) {
        if (!nextProps.show) {
            // trigger hide
            this.setState({
                reload: true
            });

            return
        }

        if (!nextProps.namespace.Name) {
            // invalid namespace
            return
        }

        // show modal
        if (this.state.reload) {
            // make copy
            let namespace = JSON.parse(JSON.stringify(nextProps.namespace));

            // set to state
            this.setState({
                namespace: namespace,
                reload: false
            });
        }
    }

    kubernetesSettingsConfig() {
        let ret = [];

        if (this.props.config.Kubernetes.Namespace.Settings) {
            ret = this.props.config.Kubernetes.Namespace.Settings;
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
                                        <input className="form-control" type="text" value={this.state.namespace.Name} readOnly />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsDescription" className="inputRg">Description</label>
                                        <input type="text" name="nsDescription" id="inputNsDescription" className="form-control" placeholder="Description" value={this.getNamespaceItem("Description")} onChange={this.handleNamespaceInputChange.bind(this, "Description")} />
                                    </div>

                                    {this.kubernetesSettingsConfig().map((setting, value) =>
                                        <K8sNamespaceFormElement setting={setting} value={this.getNamespaceSettingItem(setting.Name)} onchange={this.handleNamespaceSettingInputChange.bind(this, setting.Name)} />
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

export default K8sNamespaceModalEdit;

