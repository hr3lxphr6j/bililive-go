import { Modal, Input } from 'antd';
import React from 'react';
import API from '../../utils/api';

const api = new API();

interface Props {
    refresh?: any
}

class AddRoomDialog extends React.Component<Props> {
    state = {
        ModalText: '请输入直播间的URL地址',
        visible: false,
        confirmLoading: false,
        textView: ''
    };

    showModal = () => {
        this.setState({
            visible: true,
        });
    };

    handleOk = () => {
        this.setState({
            ModalText: '正在添加直播间......',
            confirmLoading: true,
        });
        
        api.addNewRoom(this.state.textView)
            .then((rsp) => {
                this.setState({
                    visible: false,
                    confirmLoading: false,
                });
                this.props.refresh();
            })
    };

    handleCancel = () => {
        console.log('Clicked cancel button');
        this.setState({
            visible: false,
        });
    };

    textChange = (e: any) =>{
        this.setState({
            textView: e.target.value
        })
    }

    render() {
        const { visible, confirmLoading, ModalText } = this.state;
        return (
            <div>
                <Modal
                    title="添加直播间"
                    visible={visible}
                    onOk={this.handleOk}
                    confirmLoading={confirmLoading}
                    onCancel={this.handleCancel}>
                    <p>{ModalText}</p>
                    <Input size="large" placeholder="https://" onChange={this.textChange}/>
                </Modal>
            </div>
        );
    }
}

export default AddRoomDialog;
