import React from "react";
import API from '../../utils/api';
import { PageHeader } from 'antd';
import { Descriptions } from 'antd';

const api = new API();

interface Props {

}

interface IState {
    appName: string
    appVersion: string
    buildTime: string
    gitHash: string
    pid: string
    platform: string
    goVersion: string
}

class LiveInfo extends React.Component<Props, IState> {

    constructor(props: Props) {
        super(props);
        this.state = {
            appName: "",
            appVersion: "",
            buildTime: "",
            gitHash: "",
            pid: "",
            platform: "",
            goVersion: ""
        }
    }

    componentDidMount() {
        api.getLiveInfo()
            .then((rsp: any) => {
                const data = rsp.data;
                this.setState({
                    appName: data.app_name,
                    appVersion: data.app_version,
                    buildTime: data.build_time,
                    gitHash: data.git_hash,
                    pid: data.pid,
                    platform: data.platform,
                    goVersion: data.go_version
                })
            })
    }

    render() {
        return (
            <div>
                <div style={{ backgroundColor: '#F5F5F5', }}>
                    <PageHeader
                        ghost={false}
                        title="系统状态"
                        subTitle="System Info">
                    </PageHeader>
                </div>
                <Descriptions bordered>
                    <Descriptions.Item label="App Name">{this.state.appName}</Descriptions.Item>
                    <Descriptions.Item label="App Version">{this.state.appVersion}</Descriptions.Item>
                    <Descriptions.Item label="Build Time">{this.state.buildTime}</Descriptions.Item>
                    <Descriptions.Item label="Pid">{this.state.pid}</Descriptions.Item>
                    <Descriptions.Item label="Platform">{this.state.platform}</Descriptions.Item>
                    <Descriptions.Item label="Go Version">{this.state.goVersion}</Descriptions.Item>
                    <Descriptions.Item label="Git Hash">{this.state.gitHash}</Descriptions.Item>
                </Descriptions>
            </div>
        )
    }
}

export default LiveInfo;