import React from 'react';
import BaseComponent from '../base';

class NamespaceDelete extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            buttonState: "",
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
            buttonState: "disabled",
            buttonText: "Deleting...",
        });

        let jqxhr = this.ajax({
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
                buttonState: "",
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
        if (this.state.buttonState !== "") {
            return this.state.buttonState;
        }

        if (this.state.confirmNamespace !== this.props.namespace.name) {
            return "disabled";
        }
    }

    render() {
        return (
            <div>
                <form method="post">
                    <div className="modal fade" id="deleteQuestion" tabIndex="-1" role="dialog" aria-labelledby="deleteQuestion">
                        <div className="modal-dialog" role="document">
                            <div className="modal-content">
                                <div className="modal-header">
                                    <h5 className="modal-title" id="exampleModalLabel">Delete namespace</h5>
                                    <button type="button" className="close" data-dismiss="modal" aria-label="Close">
                                        <span aria-hidden="true">&times;</span>
                                    </button>
                                </div>
                                <div className="modal-body">
                                    <div className="row">
                                        <div className="col">
                                            Do you really want to delete namespace <strong className="k8s-namespace">{this.props.namespace.name}</strong>?
                                        </div>
                                    </div>
                                    <div className="row">
                                        <div className="col">
                                            <input type="text" id="inputNsDeleteConfirm" className="form-control" placeholder="Enter namespace for confirmation" required value={this.getValue("confirmNamespace")} onChange={this.setValue.bind(this, "confirmNamespace")} />
                                        </div>
                                    </div>
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-primary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="submit" className="btn btn-secondary bnt-k8s-namespace-delete" disabled={this.renderButtonState()} onClick={this.deleteNamespace.bind(this)}>{this.state.buttonText}</button>
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

