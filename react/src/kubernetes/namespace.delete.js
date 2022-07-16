import React from 'react';
import BaseComponent from '../base';

class NamespaceDelete extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            requestRunning: false,
            buttonText: "Delete namespace",
            confirmNamespace: ""
        };
    }

    deleteNamespace(e) {
        e.preventDefault();
        e.stopPropagation();

        if (!this.props.namespace) {
            return
        }

        let oldButtonText = this.state.buttonText;
        this.setState({
            requestRunning: true,
            buttonText: "Deleting...",
        });

        this.ajax({
            type: 'DELETE',
            url: "/_webapi/kubernetes/namespace/" + encodeURI(this.props.namespace.name)
        }).done(() => {
            this.setState({
                confirmNamespace: ""
            });

            if (this.props.callback) {
                this.props.callback()
            }
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });
    }

    componentWillReceiveProps(nextProps) {
        if (!this.props.namespace || this.props.namespace.name !== nextProps.namespace.name) {
            this.setState({
                confirmNamespace: ""
            });
        }
    }

    renderButtonState() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else if (this.state.confirmNamespace !== this.props.namespace.name) {
            state = "disabled"
        }

        return state
    }

    render() {
        return (
            <div>
                <form method="post">
                    <div className="modal fade" id="deleteQuestion" tabIndex="-1" role="dialog"
                         aria-labelledby="deleteQuestion">
                        <div className="modal-dialog" role="document">
                            <div className="modal-content">
                                <div className="modal-header">
                                    <h5 className="modal-title" id="exampleModalLabel">Delete namespace</h5>
                                    <button type="button" className="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                                </div>
                                <div className="modal-body">
                                    <div className="row">
                                        <div className="col">
                                            Do you really want to delete namespace <strong
                                            className="k8s-namespace">{this.props.namespace.name}</strong>?
                                        </div>
                                    </div>
                                    <div className="row">
                                        <div className="col">
                                            <input type="text" id="inputNsDeleteConfirm" className="form-control"
                                                   placeholder="Enter namespace for confirmation" required
                                                   value={this.getValue("confirmNamespace")}
                                                   onChange={this.setValue.bind(this, "confirmNamespace")}/>
                                        </div>
                                    </div>
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-primary bnt-k8s-namespace-cancel"
                                            data-bs-dismiss="modal">Cancel
                                    </button>
                                    <button type="submit" className="btn btn-danger bnt-k8s-namespace-delete"
                                            disabled={this.renderButtonState()}
                                            onClick={this.deleteNamespace.bind(this)}>{this.state.buttonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default NamespaceDelete;

