import React from "react";
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-javascript';
import 'prismjs/themes/prism.css'; //Example style, you can use another
import API from '../../utils/api';
import { Collapse, Input, Icon, InputNumber, Button, Select } from "antd";
import './config-info.css';
const { Panel } = Collapse;
const { Option } = Select;
const { TextArea } = Input;
const api = new API();

interface Props {

}

interface IState {
  config: any
}

class ConfigInfo extends React.Component<Props, IState> {

  constructor(props: Props) {
    super(props);
    this.state = {
      config: null,
    }
  }

  componentDidMount(): void {
    api.getConfigInfo()
      .then((rsp: any) => {
        this.setState({
          config: rsp.config
        });
      });
  }

  /**
     * 保存设置至config文件
     */
  onSettingSave = () => {
    api.saveRawConfig({ config: this.state.config })
      .then((rsp: any) => {
        if (rsp.err_no === 0) {
          alert("设置保存成功");
        } else {
          alert("Server Error!");
        }
      })
  }

  renderConfigString(value: string): JSX.Element {
    return <Input defaultValue={value} />;
  }

  renderConfigNumber(value: number): JSX.Element {
    return <InputNumber defaultValue={value} />;
  }

  renderConfigArray(value: any[]): JSX.Element[] {
    const ret: JSX.Element[] = [];
    value.forEach((val, i) => {
      ret.push(this.renderConfigObj(val))
    });
    return ret;
  }

  renderConfigObj(obj: any): JSX.Element {
    const children = [];
    for (const [key, value] of Object.entries(obj)) {
      if (key === "RPC" || key === "LiveRooms") {
        continue;
      }
      let child: JSX.Element | JSX.Element[] | null = null;
      switch (typeof value) {
        case "string":
          child = this.renderConfigString(value);
          break;
        case "number":
          child = this.renderConfigNumber(value);
          break;
        case "boolean":
          continue;
        case "object":
          if (Array.isArray(value)) {
            child = this.renderConfigArray(value);
          } else {
            child = this.renderConfigObj(value);
          }
          break;
        default:
          console.error(`未知设置： ${key}:${JSON.stringify(value)}`);
          continue;
      }
      if (child !== null) {
        children.push(
          <Panel
            header={
              <div>
                {key}
                <Button
                  style={{
                    marginLeft: "16px"
                  }}
                  icon="delete"
                  type="danger"
                  onClick={(e) => {
                    e.preventDefault();
                    delete obj[key];
                    this.setState(this.state.config);
                  }}
                />
              </div>
            }
            key={key}
          >
            {child}
          </Panel >
        );
      }
    }

    let inputRef: Input | null = null;
    return <div><Collapse defaultActiveKey={Object.keys(obj)}>
      {children}
    </Collapse>

      <Input.Group
        compact
        style={{
          display: "flex",
          margin: "16px auto",
          flexDirection: "row",
          alignItems: "center",
        }}
      >
        增加 key:
        <Input
          style={{
            width: "50%",
            marginLeft: "16px",
          }}
          ref={ref => {
            inputRef = ref;
          }}
        />
        <Select
          defaultValue="string"
          style={{
            width: 90,
            marginRight: 12,
          }}
        >
          <Option value="string">
            字符串
          </Option>
          <Option value="number">
            数字
          </Option>
          <Option value="boolean">
            布尔值
          </Option>
        </Select>
        <Button
          icon="file-add"
          type="primary"
          onClick={(e) => {
            if (!inputRef || !inputRef.state.value) {
              return;
            }
            obj[inputRef.state.value] = "";
            this.setState(this.state.config);
          }}
        /></Input.Group></div>;
  }

  render() {
    if (this.state.config === null) {
      return <div>loading...</div>;
    }
    // return this.renderConfigObj(this.state.config);
    return <div>
      <Editor
        value={this.state.config}
        onValueChange={code => this.setState({ config: code })}
        highlight={code => {
          const ret = highlight(code, languages.js, "js");
          return ret;
        }}
        padding={10}
        style={{
          fontFamily: '"Fira code", "Fira Mono", monospace',
          fontSize: 12,
        }}
      />
      <TextArea
        style={{
          minHeight: 680,
        }}
        defaultValue={this.state.config}
        onChange={e => this.setState({ config: e.target.value })}
      />
      <Button type="default" onClick={this.onSettingSave}>保存设置</Button>
    </div>
  }
}

export default ConfigInfo;