import React from 'react';
import './App.css';
import 'antd/dist/antd.css';
import { Switch, Route } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/es/locale/zh_CN'
import RootLayout from './component/layout/index';
import LiveList from './component/live-list/index';
import LiveInfo from './component/live-info/index';

const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN}>
      <RootLayout>
        <Switch>
          <Route path="/liveInfo" component={LiveInfo}></Route>
          <Route path="/" component={LiveList}></Route>
        </Switch>
      </RootLayout>
    </ConfigProvider>

  );
}

export default App;
