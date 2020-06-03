import React from 'react';
import { HashRouter as Router, Link } from 'react-router-dom';
import { Layout, Menu, Icon } from 'antd';

const { SubMenu } = Menu;
const { Header, Content, Sider } = Layout;

class RootLayout extends React.Component {
    render() {
        return (
            <Layout className="all-layout">
                <Header className="header">
                    <h3 className="logo-text">Bililive-go</h3>
                </Header>
                <Layout>
                    <Router>
                        <Sider width={200} style={{ background: '#fff' }}>
                            <Menu
                                mode="inline"
                                defaultSelectedKeys={['1']}
                                defaultOpenKeys={['sub1']}
                                style={{ height: '100%', borderRight: 0 }}
                            >
                                <SubMenu
                                    key="sub1"
                                    title={
                                        <span>
                                            <Icon type="monitor" />
                                            LiveClient
                                        </span>
                                    }
                                >
                                    <Menu.Item key="1"><Link to="/">监控列表</Link></Menu.Item>
                                    <Menu.Item key="2"><Link to="/liveInfo">系统状态</Link></Menu.Item>
                                </SubMenu>
                            </Menu>
                        </Sider>
                        <Layout style={{ padding: '0px 24px' }}>
                            <Content
                                style={{
                                    background: '#fff',
                                    padding: 24,
                                    margin: 0,
                                    minHeight: 280,
                                }}>
                                {this.props.children}
                            </Content>
                        </Layout>
                    </Router>
                </Layout>
            </Layout>
        )
    }
}

export default RootLayout;
