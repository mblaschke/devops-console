import React from 'react';
import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import Breadcrumb from "./Breadcrumb";

class K8sAccess extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            kubeconfig: false
        };
    }

    loadKubeconfig() {
        let jqxhr = this.ajax({
            type: "GET",
            url: '/api/kubernetes/kubeconfig'
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

        return (
            <div>
                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-server"></i>
                        Kubernetes Access
                    </div>
                    <div className="card-body card-body-table spinner-area">
                        <textarea className="kubeconfig" readOnly="readOnly" value={this.state.kubeconfig}/>
                        <div className="d-flex justify-content-end">
                            <a href="/api/kubernetes/kubeconfig" download="kubeconfig.json" className="btn btn-secondary btn-lg active" role="button" aria-pressed="true">Download kubeconfig</a>
                        </div>
                    </div>
                    <div className="card-footer small text-muted"></div>
                </div>
            </div>
        );
    }
}

export default K8sAccess;

