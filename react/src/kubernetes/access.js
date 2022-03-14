import React from 'react';
import BaseComponent from '../base';
import Spinner from '../spinner';
import Breadcrumb from '../breadcrumb';

class Access extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            kubeconfig: false
        };
    }

    loadKubeconfig() {
        this.ajax({
            type: "GET",
            url: '/_webapi/kubernetes/kubeconfig'
        }).done((jqxhr) => {
            this.setState({
                isStartup: false,
                kubeconfig: jqxhr
            })
        });
    }

    componentDidMount() {
        this.loadKubeconfig();
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

        let kubeconfigs = {}
        if (this.state.kubeconfig) {
            kubeconfigs = this.state.kubeconfig
        }

        return (
            <div>
                <Breadcrumb/>

                {Object.keys(kubeconfigs).map((name) =>
                    <div className="card mb-3">
                        <div className="card-header">
                            <i className="fas fa-sign-in-alt"></i>
                            {kubeconfigs[name].name}
                        </div>
                        <div className="card-body card-body-table spinner-area">
                            <div className="form-group">
                                <label htmlFor="textareaKubeconfig">{kubeconfigs[name].description}</label>
                                <textarea id="textareaKubeconfig" className="kubeconfig" readOnly="readOnly"
                                          value={kubeconfigs[name].content}/>
                                <small className="form-text text-muted">Save as ~/.kube/config</small>
                                <div className="d-flex justify-content-end">
                                    <a href={"/_webapi/kubernetes/kubeconfig/" + name} download="kubeconfig.json"
                                       className="btn btn-secondary btn-lg active" role="button" aria-pressed="true">Download
                                        kubeconfig</a>
                                </div>
                            </div>
                        </div>
                        <div className="card-footer small text-muted"></div>
                    </div>
                )}
            </div>
        );
    }
}

export default Access;

