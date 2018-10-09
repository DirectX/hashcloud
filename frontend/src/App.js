import React, { Component } from 'react';
import { Alert, Button, Container, Form, FormGroup, CustomInput, Nav, NavItem, NavLink, Row, Col } from 'reactstrap';
import Web3 from 'web3';
import FileSaver from 'file-saver';
import { sha256Hash } from './sha256';
import './App.css';

function humanFileSize(bytes, si) {
  var thresh = si ? 1000 : 1024;

  if(Math.abs(bytes) < thresh) {
    return bytes + ' B';
  }
  var units = si
    ? ['kB','MB','GB','TB','PB','EB','ZB','YB']
    : ['KiB','MiB','GiB','TiB','PiB','EiB','ZiB','YiB'];
  var u = -1;
  do {
    bytes /= thresh;
    ++u;
  } while(Math.abs(bytes) >= thresh && u < units.length - 1);
  return bytes.toFixed(1)+' '+units[u];
}

class App extends Component {
  constructor(props) {
    super(props);

    if (!window.web3) {
      this.state = {
        account: '',
        balance: '0',
        metamaskWarningOpen: true,
      };

      return;
    }

    this.state = {
      account: '',
      balance: '0',
      metamaskWarningOpen: false,
      web3js: new Web3(window.web3.currentProvider),
      files: [],
      newFiles: [],
      result: [],
      uploaded: false,
    };

    this.onSelectFiles = this.onSelectFiles.bind(this);
    this.onUpload = this.onUpload.bind(this);
    this.onCancel = this.onCancel.bind(this);
    this.onDownload = this.onDownload.bind(this);

    this.state.web3js.eth.net.getNetworkType().then((networkName) => {
      if (networkName !== 'rinkeby') {
        this.setState({ metamaskWarningOpen: true });
      } else {
        window.web3.currentProvider.publicConfigStore.on('update', () => {
          this.setMetaMaskAccount();
        });
        this.setMetaMaskAccount();
      }
    });
  }

  formatPrice(weiPriceString) {
    if (this.state.web3js) {
      return parseFloat(this.state.web3js.utils.fromWei(weiPriceString)).toFixed(3);
    } else {
      return '0';
    }
  }

  filePanelClass(file) {
    if (file.uploaded) {
      if (file.stored)
        return "bg-success text-white"
      else
        return "bg-secondary text-white"
    }

    return "";
  }

  async setMetaMaskAccount() {
    let self = this;

    let accounts = await this.state.web3js.eth.getAccounts();
    if (accounts.length === 0) {
      this.setState({ metamaskWarningOpen: true });
      return;
    }

    let account = accounts[0];

    if (account && this.state.account !== account) {
      let balance = await this.state.web3js.eth.getBalance(account);
      self.setState({ account: accounts[0], balance: balance.toString() });

      this.loadUserFiles();
    }
  }

  async loadUserFiles() {
    const account = this.state.account;

    let result = await fetch(`${process.env.REACT_APP_API_URL_PREFIX}/list/${account}`, {
      method: 'GET'
    });

    let files = await result.json();

    this.setState({ files: files });
  }

  async onSelectFiles(event) {
    const fileList = event.target.files;
    let promises = [];
    let totalSize = 0;
    let fileCount = 0;

    for (let index = 0; index < fileList.length; index++) {
      const file = fileList[index];
      
      promises.push(new Promise((resolve, reject) => {
        if (file.size > 20 * 1024 * 1024) {
          resolve({ index: index, hash: '', file: file, oversize: true });
        } else {
          let reader = new FileReader();
          
          reader.onload = function(loadedEvent) {
            const arrayBuffer = loadedEvent.target.result;
            const hash = sha256Hash(arrayBuffer);
            totalSize += file.size;
            fileCount++;

            resolve({ hash: hash, file: file });
          };
          reader.readAsArrayBuffer(file);
        }
      }));
    }

    let files = await Promise.all(promises);

    this.setState({ newFiles: files, totalSize: totalSize, fileCount: fileCount });
  }

  async onUpload(event) {
    try {
      let files = this.state.newFiles;
      let hashes = [];

      files.map(file => {
        hashes.push(file.hash);
      });
      
      const allHashString = hashes.join('+');
      const dataHash = this.state.web3js.utils.sha3(allHashString);
      const signature = await this.state.web3js.eth.personal.sign(dataHash, this.state.account);

      let data = new FormData()
      data.append('meta', JSON.stringify({ owner: this.state.account, signature: signature }));
      files.map(file => {
        data.append('files', file.file, file.file.name);
      });

      let result = await fetch(`${process.env.REACT_APP_API_URL_PREFIX}/upload`, {
        method: 'POST',
        body: data,
      });

      let resultJson = await result.json();

      files.map(file => {
        file.uploaded = true;
        file.stored = resultJson.includes(file.hash);
      });

      this.setState({ newFiles: files, result: resultJson, uploaded: true });
    } catch(err) {
      console.log(err);
    }
  }

  async onCancel(event) {
    if (this.state.result.length > 0)
      this.loadUserFiles();

    this.setState({ newFiles: [], result: [], uploaded: false });
  }

  async onDownload(hash, filename) {
    try {
      const dataHash = this.state.web3js.utils.sha3(hash);
      const signature = await this.state.web3js.eth.personal.sign(dataHash, this.state.account);

      let data = new FormData()
      data.append('meta', JSON.stringify({ owner: this.state.account, signature: signature }));

      let response = await fetch(`${process.env.REACT_APP_API_URL_PREFIX}/download/${hash}`, {
        method: 'POST',
        body: data,
      });

      let blob = await response.blob();

      FileSaver.saveAs(blob, filename);
    } catch (err) {
      alert('Unathorized');
    }
  }

  render() {
    return (
      <div className="App">
        <Nav className="navbar navbar-expand-md navbar-dark bg-dark fixed-top">
          <NavLink className="navbar-brand" href="#">Hash Cloud</NavLink>
          <button className="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarsExampleDefault" aria-controls="navbarsExampleDefault" aria-expanded="false" aria-label="Toggle navigation">
            <span className="navbar-toggler-icon"></span>
          </button>
          <div className="collapse navbar-collapse" id="navbarsExampleDefault">
            <ul className="navbar-nav mr-auto">
              <li className="nav-item active" hidden={true}>
                <NavLink href="#">Home <span className="sr-only">(current)</span></NavLink>
              </li>
            </ul>
            <NavItem className="text-white mr-4">
              Account: <strong>{this.state.account}</strong>
            </NavItem>
            <NavItem className="text-white mr-3" hidden={true}>
              ETH Balance: <strong>{this.formatPrice(this.state.balance)}</strong>&nbsp;ETH
            </NavItem>
          </div>
        </Nav>

        <Container>
          <Alert color="info" isOpen={this.state.metamaskWarningOpen} toggle={this.onDismissMetamaskInfo}>
            Please unlock MetaMask account and select Rinkeby test network
          </Alert>
          <h1 className="mx-3">Cloud Storage</h1>
          <div hidden={this.state.newFiles.length !== 0} className="files-panel">
            <Form className="mx-3">
              <FormGroup>
                <CustomInput type="file" id="fileBrowser" name="file" label="Select and upload new files..." onChange={this.onSelectFiles} multiple />
              </FormGroup>
            </Form>
          </div>
          <div hidden={this.state.newFiles.length === 0} className="files-panel">
            {this.state.newFiles.map(file => {
              return <Container className={"file-panel p-3 shadow " + this.filePanelClass(file)}>
                <Row className="lead align-items-center">
                  <Col className="col-7 text-truncate" title={file.file.name}><span className="font-weight-bold">{file.file.name}</span></Col>
                  <Col className="col-3 text-muted">{file.file.type}</Col>
                  <Col className={"col-2 text-right " + (file.oversize ? "font-weight-bold text-danger" : "")}>{humanFileSize(file.file.size, true)}</Col>
                </Row>
              </Container>
            }, this)}
          </div>
          <div hidden={this.state.uploaded || this.state.newFiles.length === 0} className="files-panel">
            <h2 className="mx-3">Summary</h2>
            <div className="container lead">
              <div className="row justify-content-start">
                <div className="col-6 text-right">
                  File owner
                </div>
                <div className="col-6">
                  {this.state.account}
                </div>
              </div>
              <div className="row justify-content-start">
                <div className="col-6 text-right">
                  File Count
                </div>
                <div className="col-6">
                  {this.state.fileCount}
                </div>
              </div>
              <div className="row justify-content-start">
                <div className="col-6 text-right">
                  Total Size
                </div>
                <div className="col-6">
                  {humanFileSize(this.state.totalSize, true)}
                </div>
              </div>
            </div>
            <br />
            <div className="text-center">
              <Button className="mr-2" color="success" size="lg" onClick={this.onUpload}>Sign and Upload&hellip;</Button>
              <Button color="secondary" size="lg" onClick={this.onCancel}>Cancel</Button>
            </div>
          </div>
          <div hidden={!this.state.uploaded || this.state.newFiles.length === 0} className="files-panel">
            <div className="text-center">
              <Button color="primary" size="lg" onClick={this.onCancel}>Done</Button>
            </div>
          </div>
          <div className="files-panel">
            <h2 className="mx-3">My Files</h2>
            {this.state.files.map(file => {
              return <Container className="file-panel p-3 shadow">
                <Row className="lead align-items-center">
                  <Col className="col-8 text-truncate" title={file.filename}><span className="font-weight-bold">{file.filename}</span><br />{file.hash}</Col>
                  <Col className="col-2 text-right">{humanFileSize(file.contentSize, true)}</Col>
                  <Col className="col-2 text-right"><Button color="primary" outline onClick={() => this.onDownload(file.hash, file.filename)}>Download...</Button></Col>
                </Row>
              </Container>
            }, this)}
          </div>
        </Container>
      </div>
    );
  }
}

export default App;