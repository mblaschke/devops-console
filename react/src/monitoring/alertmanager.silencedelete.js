import React from 'react';
import BaseComponent from '../base';

class AlertmanagerSilencedelete extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            buttonState: "",
            buttonText: "Delete silence",
            confirmNamespace: ""
        };
    }

    confirm(e) {
        e.preventDefault();
        e.stopPropagation();

        if (!this.props.silence) {
            return
        }

        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Deleting...",
        });

        let jqxhr = this.ajax({
            type: 'DELETE',
            url: "/_webapi/alertmanager/" + encodeURI(this.props.instance) + "/silence/" + encodeURI(this.props.silence.id)
        }).done(() => {
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

    handleConfirmNamespace(event) {
        this.setState({
            confirmNamespace: event.target.value
        });
    }

    render() {
        if (!this.props.silence) {
            return <div></div>;
        }

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
                                            Do you really want to delete silence?
                                            <strong>{this.props.silence.description}</strong>
                                        </div>
                                    </div>
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-primary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="submit" className="btn btn-secondary bnt-k8s-namespace-delete" onClick={this.confirm.bind(this)}>{this.state.buttonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default AlertmanagerSilencedelete;

