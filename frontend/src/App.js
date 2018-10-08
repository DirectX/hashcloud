import React, { Component } from 'react';
import { Alert, Button, Container, Form, FormGroup, CustomInput, Nav, NavItem, NavLink, Row, Col } from 'reactstrap';
import Web3 from 'web3';
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
    };

    this.onSelectFiles = this.onSelectFiles.bind(this);
    this.onUpload = this.onUpload.bind(this);

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
    }
  }

  async onSelectFiles(event) {
    var fileList = event.target.files;
    var promises = [];
    var totalSize = 0;

    for (let index = 0; index < fileList.length; index++) {
      const file = fileList[index];
      
      promises.push(new Promise((resolve, reject) => {
        var reader = new FileReader();
        
        reader.onload = function(loadedEvent) {
          const arrayBuffer = loadedEvent.target.result;
          const hash = sha256Hash(arrayBuffer);
          totalSize += file.size;

          resolve({ index: index, hash: hash, file: file });
        };
        reader.readAsArrayBuffer(file);
      }));
    }

    let files = await Promise.all(promises);

    this.setState({ files: files, totalSize: totalSize });
  }

  async onUpload(event) {
    try {
      const dataHash = this.state.web3js.utils.sha3("Some text");
      const signature = await this.state.web3js.eth.personal.sign(dataHash, this.state.account);
      //const recoveredSender = await this.state.web3js.eth.personal.ecRecover(dataHash, signature) === this.state.account.toLowerCase();
      //console.log(signature, recoveredSender);

      var data = new FormData()
      for (const file of this.state.files) {
        data.append('files', file.file, file.file.name);
      }

      fetch('/upload', {
        method: 'POST',
        body: data
      });
    
      this.setState({ signature: signature });
    } catch(err) {
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
              <li className="nav-item active">
                <NavLink href="#">Home <span className="sr-only">(current)</span></NavLink>
              </li>
            </ul>
            <NavItem className="text-white mr-4">
              Account: <strong>{this.state.account}</strong>
            </NavItem>
            <NavItem className="text-white mr-3">
              ETH Balance: <strong>{this.formatPrice(this.state.balance)}</strong>&nbsp;ETH
            </NavItem>
          </div>
        </Nav>

        <Container>
          <Alert color="info" isOpen={this.state.metamaskWarningOpen} toggle={this.onDismissMetamaskInfo}>
            Please unlock MetaMask account and select Rinkeby test network
          </Alert>
          <h1>Upload Files</h1>
          <br />
          <Form>
            <FormGroup>
              <CustomInput type="file" id="fileBrowser" name="file" label="Select files..." onChange={this.onSelectFiles} multiple />
            </FormGroup>
          </Form>
          <div hidden={this.state.files.length === 0} className="files-panel">
            <h2>Loaded Files</h2>
            {this.state.files.map(function (file) {
              return <Container className="file-panel p-3 shadow">
                <Row className="align-items-center">
                  <Col className="lead"><span className="font-weight-bold">{file.file.name}</span></Col>
                  <Col className="lead">{file.hash}</Col>
                  <Col className="lead col-1">{humanFileSize(file.file.size, true)}</Col>
                </Row>
              </Container>
            }, this)}
          </div>
          <div hidden={this.state.files.length === 0} className="files-panel">
            <h2>Summary</h2>
            <br />
            <div className="text-center" onClick={this.onUpload}><Button color="success" size="lg">Upload</Button></div>
          </div>
        </Container>
      </div>
    );
  }
}

export default App;