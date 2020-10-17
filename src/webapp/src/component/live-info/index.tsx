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
                this.setState({
                    appName: rsp.app_name,
                    appVersion: rsp.app_version,
                    buildTime: rsp.build_time,
                    gitHash: rsp.git_hash,
                    pid: rsp.pid,
                    platform: rsp.platform,
                    goVersion: rsp.go_version
                })
            })
    }

    render() {
        return (
            <div>
                <div style={{ backgroundColor: '#F5F5F5', }}>
                    <PageHeader
                        ghost={false}
                        title="系统信息">
                    </PageHeader>
                </div>
                <Descriptions bordered>
                    <Descriptions.Item label="App名称">{this.state.appName}</Descriptions.Item>
                    <Descriptions.Item label="App版本">{this.state.appVersion}</Descriptions.Item>
                    <Descriptions.Item label="编译时间">{this.state.buildTime}</Descriptions.Item>
                    <Descriptions.Item label="Pid">{this.state.pid}</Descriptions.Item>
                    <Descriptions.Item label="平台">{this.state.platform}</Descriptions.Item>
                    <Descriptions.Item label="Go 版本">{this.state.goVersion}</Descriptions.Item>
                    <Descriptions.Item label="Git 提交版本">{this.state.gitHash}</Descriptions.Item>
                </Descriptions>
            </div>
        )
    }
}

export default LiveInfo;
