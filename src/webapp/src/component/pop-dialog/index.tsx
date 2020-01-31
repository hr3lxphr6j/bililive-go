import { Popconfirm, Icon } from 'antd';
import React from 'react';

interface DialogContent{
    title: string,
    onConfirm?: (e?: React.MouseEvent<HTMLElement>) => void
}

class PopDialog extends React.Component<DialogContent> {
    render() {
        return (
            <Popconfirm
                title={this.props.title}
                icon={<Icon type="question-circle-o" style={{ color: 'red' }} />}
                onConfirm={this.props.onConfirm}>
                {this.props.children}
            </Popconfirm>
        );
    }
}

export default PopDialog;
