import React from 'react';
import {
  BrowserRouter as Router,
  NavLink,
  Switch,
  Route
} from "react-router-dom";
import Moment from 'react-moment';
import './App.css';
import {Button, Form, FormGroup, Label, Input, Table, Container, Row, Col, NavLink as RSNavLink, Navbar, Nav, NavbarToggler, NavItem, NavbarBrand, Collapse} from 'reactstrap';
import * as numeral from 'numeral';
import VisNetworkReactComponent from "vis-network-react";

const customShapeRenderer = ({
  ctx,
  x,
  y,
  state: { selected, hover },
  style,
  label,
}) => {
  const splittedLabel = label.split("---");
  ctx.save();
  ctx.restore();
  const type = splittedLabel[0];
  let hash = `Hash: ${splittedLabel[1].substr(0,8)}...${splittedLabel[1].substr(-8)}`;
  let height = `Height: ${splittedLabel[2]}`;

  if(splittedLabel[1] === "exp") {
    height = '';
    hash = `... expand ${splittedLabel[3]} more ...`;
    style.color = "#c0c0c0";
  }

  const r = 5;

  const labelWidth = ctx.measureText(height).width;
  const valueWidth = ctx.measureText(hash).width;

  const wPadding = 10;
  let hPadding = 10;

  const w = 200;
  const h = 60;
  const drawNode = () => {
    const r2d = Math.PI / 180;
    if (w - 2 * r < 0) {
      r = w / 2;
    } //ensure that the radius isn't too large for x
    if (h - 2 * r < 0) {
      r = h / 2;
    } //ensure that the radius isn't too large for y

    const top = y - h / 2;
    const left = x - w / 2;

    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(left + r, top);
    ctx.lineTo(left + w - r, top);
    ctx.arc(left + w - r, top + r, r, r2d * 270, r2d * 360, false);
    ctx.lineTo(left + w, top + h - r);
    ctx.arc(left + w - r, top + h - r, r, 0, r2d * 90, false);
    ctx.lineTo(left + r, top + h);
    ctx.arc(left + r, top + h - r, r, r2d * 90, r2d * 180, false);
    ctx.lineTo(left, top + r);
    ctx.arc(left + r, top + r, r, r2d * 180, r2d * 270, false);
    ctx.closePath();
    ctx.save();
    ctx.fillStyle = style.color || "#56CCF2";
    ctx.fill();
    ctx.strokeStyle = "#000000";
    ctx.stroke();

    // label text
    ctx.font = "normal 12px sans-serif";
    ctx.fillStyle = "grey";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    let textHeight1 = 12;
    ctx.fillText(
      height,
      left + w / 2,
      top + textHeight1 + hPadding,
      w
    );

    // value text
    ctx.font = "bold 14px sans-serif";
    ctx.fillStyle = "black";
    const textHeight2 = 12;

    if(splittedLabel[1] === "exp") {
      hPadding = 0;
    }
    ctx.fillText(hash, left + w / 2, top + h / 2 + hPadding, w);
  };

  ctx.save();
  ctx.restore();
  return {
    drawNode,
    nodeDimensions: { width: w, height: h },
  };
};

const ShortHash = (props) => {
  var len = props.len || 8;
  return <>{props.hash.substring(0,len)}...{props.hash.substr(0-len)}</>
}

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpen: false,
      reorgs: [],
      coins: [],
      minRemoved: 6,
      coinTicker: "",
      onlyWithDoubleSpends: false,
      detailID: null,
      detail: null,
      removedExpansion: 0,
      addedExpansion: 0,
    }
    this.refreshDetail = this.refreshDetail.bind(this);
    this.refresh = this.refresh.bind(this);
    this.chainGraphNodeClick = this.chainGraphNodeClick.bind(this);
    this.toggle = this.toggle.bind(this);
  }
  
  componentDidMount() {
    fetch("https://pooldetective.org/api/public/coins").then(r=>r.json()).then((r)=>{
      var coins = r.sort(( a, b ) => {
        if ( a.name < b.name ){
          return -1;
        }
        if ( a.name > b.name ){
          return 1;
        }
        return 0;
      })
      this.setState({coins:coins}, this.refresh)}
    );
  }

  chainGraphNodeClick(params) {
    if(params.nodes && params.nodes[0]) { 
      let nodeParts = params.nodes[0].split('---');
      if(nodeParts.length == 3) {
        if(nodeParts[0] === "exp") {
          params.event.preventDefault();
          if(nodeParts[1] === "add") {
            this.setState({addedExpansion : this.state.addedExpansion+10})
          }
          if(nodeParts[1] === "rem") {
            this.setState({removedExpansion : this.state.removedExpansion+10})
          }
        }
      }
    }
  }

  refreshDetail() {
    fetch(`https://pooldetective.org/api/public/reorg/${this.state.detailID}`).then(r=>r.json()).then((r)=>{this.setState({detail:r})});
  }

  refresh() {
    fetch(`https://pooldetective.org/api/public/reorgs?minRemoved=${this.state.minRemoved}&ticker=${this.state.coinTicker}${this.state.onlyWithDoubleSpends === true ? '&onlyWithDoubleSpends=1' : ''}`).then(r=>r.json()).then((r)=>{this.setState({reorgs:r})});
  }

  toggle() {
    this.setState({isOpen:!this.state.isOpen})
  } 

  render() {
    console.log(this.state);
  return (
    <div className="App">
      <Router>
      <Navbar fixed color="light" light expand="md">
        <NavbarBrand href="/">Reorg Tracker</NavbarBrand>
        <NavbarToggler onClick={this.toggle} />
        <Collapse isOpen={this.state.isOpen} navbar>
          <Nav className="mr-auto" navbar>
            <NavItem>
              <RSNavLink tag={NavLink} to="/">Reorgs</RSNavLink>
            </NavItem>
            <NavItem> 
              <RSNavLink tag={NavLink} to="/about">About</RSNavLink>
            </NavItem>
            {/*<NavItem>
              <RSNavLink href="https://github.com/mit-dci/pooldetective/tree/master/reorgtracker">GitHub</RSNavLink>
            </NavItem>*/}
            </Nav></Collapse></Navbar>
            <Container fluid>
              <Row>
                <Col>
                    <Switch>
                      <Route exact path="/" render={({history}) => (

                      
                      <Container><Row><Col align="left">
                        <h1>Reorgs</h1>
                        <Container>
                          <Row>
                            <Col>
                              <Form>
                                <Row form>
                                  <Col sm={6} md={3}>
                                    <FormGroup>
                                      <Label for="coinTicker">Coin:</Label>
                                      <Input type="select" name="coinTicker" value={this.state.coinTicker} onChange={(e) => { this.setState({coinTicker : e.target.value}); }}>
                                        <option value="">-- All --</option>
                                      {this.state.coins.filter((c) => c.bestHash !== undefined && c.bestHash !== "").map((c) => <option key={c.id} value={c.ticker}>{c.name} ({c.ticker})</option>)}
                                      </Input>
                                    </FormGroup>
                                  </Col>
                                  <Col sm={6} md={3}>
                                    <FormGroup>
                                      <Label for="minRemoved">Min. removed:</Label>
                                      <Input type="text" name="minRemoved" value={this.state.minRemoved} onChange={(e) => { this.setState({minRemoved :(isNaN(parseInt(e.target.value)) ? 0 : parseInt(e.target.value))}); }} />
                                    </FormGroup>
                                  </Col>
                                  <Col sm={6} md={3}>
                                    <FormGroup>
                                      <Label for="onlyDoubleSpends">Double spends:</Label>
                                      <Row style={{paddingLeft:'15px'}}>
                                        <Button name="onlyDoubleSpends" color="secondary" onClick={(e) => { this.setState({onlyWithDoubleSpends : !this.state.onlyWithDoubleSpends}); }}>{this.state.onlyWithDoubleSpends ? 'At least one' : 'Any'}</Button>
                                      </Row>
                                    </FormGroup>
                                  </Col>
                                  <Col sm={6} md={3} style={{paddingTop: '31px'}} align="right">
                                    <Button name="filter" color="primary" onClick={this.refresh}>Filter</Button>
                                  </Col>
                                </Row>
                              </Form>
                            </Col>
                          </Row>
                          <Row>
                            <Col>
                              <Table striped>
                                <thead>
                                  <tr>
                                    <th>When</th>
                                    <th>Coin</th>
                                    <th>Blocks<br/>removed</th>
                                    <th>Blocks<br/>added</th>
                                    <th>Outputs<br/>double spent</th>
                                    <th>Budish<br/>Cost</th>
                                    <th>NiceHash<br/>Cost</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {this.state.reorgs.map((r) => <tr onClick={(e) => { history.push(`/reorg/${r.id}`) }} className="doubleSpendRow" key={r.id}>
                                      <td><Moment format="L">{r.occurred}</Moment> <Moment format="LT">{r.occurred}</Moment></td>
                                      <td>{r.coinName} ({r.coinTicker})</td>
                                      <td align="center">{r.removedBlocks}</td>
                                      <td align="center">{r.addedBlocks}</td>
                                      <td align="center">{r.doubleSpentOutputs}</td>
                                      <td align="right">{r.budishCost > 0 ? numeral(r.budishCost).format('$0,0.00') : '?'}</td>
                                      <td align="right">{r.niceHashCost > 0 ? numeral(r.niceHashCost).format('$0,0.00') : '?'}</td>
                                  </tr>)}
                                </tbody>
                              </Table>
                            </Col>
                          </Row>
                        </Container>
                        </Col></Row></Container>)} />
                      
                      <Route path="/reorg/:reorgId" render={({match, history}) => {
                          var networkData = {nodes:[], edges:[]}

                          
                          if(this.state.detailID !== match.params.reorgId) {
                            this.setState({detail:null,detailID:match.params.reorgId}, this.refreshDetail)
                          }

                          if(this.state.detail === null) {
                            return <Container></Container>
                          }
                          var r = Object.assign({}, this.state.reorgs.find((r) => r.id === parseInt(match.params.reorgId)) || {}, {detail: this.state.detail});
                          
                          if(r.forkBlock !== undefined) {

                            var renderAddedBlocks = []
                            var renderRemovedBlocks = []

                            if(r.detail.addedBlocks.length > 20+this.state.addedExpansion) {
                              renderAddedBlocks = renderAddedBlocks.concat(r.detail.addedBlocks.slice(0,9+this.state.addedExpansion), [`exp---add---${r.detail.addedBlocks.length-18-this.state.addedExpansion}`], r.detail.addedBlocks.slice(r.detail.addedBlocks.length-9));
                            } else {
                              renderAddedBlocks = renderAddedBlocks.concat(r.detail.addedBlocks);
                            }

                            if(r.detail.removedBlocks.length > 20+this.state.removedExpansion) {
                              renderRemovedBlocks=renderRemovedBlocks.concat(r.detail.removedBlocks.slice(0,9+this.state.removedExpansion), [`exp---rem---${r.detail.removedBlocks.length-18-this.state.removedExpansion}`], r.detail.removedBlocks.slice(r.detail.removedBlocks.length-9));
                            } else {
                              renderRemovedBlocks=renderRemovedBlocks.concat(r.detail.removedBlocks);
                            }


                            networkData.nodes.push({id:r.forkBlock, color:'#0000ff', shape:'custom', ctxRenderer: customShapeRenderer, label:`frk---${r.forkBlock}---${r.forkBlockHeight}`})
                            renderAddedBlocks.filter((b) => b !== undefined).forEach((b,i,arr) => {
                              networkData.nodes.push({id:b, color:'#00ff00', shape:'custom', ctxRenderer: customShapeRenderer, label:`add---${b}---${r.forkBlockHeight+i+1}`})
                            });
                            renderRemovedBlocks.filter((b) => b !== undefined).forEach((b,i,arr) => {
                              networkData.nodes.push({id:b, color:'#ff0000', shape:'custom', ctxRenderer: customShapeRenderer, label:`rem---${b}---${r.forkBlockHeight+i+1}`})
                            });
                            renderAddedBlocks.filter((b) => b !== undefined).forEach((b,i,arr) => {
                              if(i == 0) { networkData.edges.push({id:'a0', from:b, to:r.forkBlock}); }
                              else { networkData.edges.push({id:`a${i}`, from:b, to:arr[i-1]}); }
                            });
                            renderRemovedBlocks.filter((b) => b !== undefined).forEach((b,i,arr) => {
                              if(i == 0) { networkData.edges.push({id:'r0', from:b, to:r.forkBlock}); }
                              else { networkData.edges.push({id:`r${i}`, from:b, to:arr[i-1]}); }
                            });
                          }


                          var visOptions = {
                            layout: {
                                hierarchical: {
                                    direction: "RL",
                                    sortMethod: "directed",
                                    nodeSpacing: 100,
                                    levelSeparation: 250
                                }
                            },
                            height:'400px',
                            edges:{arrows:{to:true}, smooth:{enabled:true, type:'cubicBezier', roundness: 1.0, forceDirection:'horizontal'}},
                            physics: {
                                enabled: false
                            },
                            interaction: {
                                dragNodes: false,
                                selectConnectedEdges: false
                            }
                        };

                        
                
                          
                          return <Container>
                            <Button onClick={(e) => {history.push('/')}} color="primary">&lt; Back</Button>
                            <h1>Reorg #{match.params.reorgId}</h1>
                            
                            <Table striped className="reorgDetail">
                              <tbody>
                                <tr>
                                  <td>Date:</td>
                                  <td><Moment format="dddd LL">{r.occurred}</Moment></td>
                                </tr>
                                <tr>
                                  <td>Time:</td>
                                  <td><Moment format="LTS">{r.occurred}</Moment></td>
                                </tr>
                                <tr>
                                  <td>Coin:</td>
                                  <td>{r.coinName} ({r.coinTicker})</td>
                                </tr>
                                <tr>
                                  <td>Fork block height:</td>
                                  <td>{r.forkBlockHeight}</td>
                                </tr>
                                <tr>
                                  <td>Fork block hash:</td>
                                  <td>{r.forkBlock}</td>
                                </tr>
                                <tr>
                                  <td>Removed:</td>
                                  <td>{r.removedBlocks} blocks / {numeral(r.removedWork).format('1.24e+7')} work</td>
                                </tr>
                                <tr>
                                  <td>Added:</td>
                                  <td>{r.addedBlocks} blocks / {numeral(r.addedWork).format('1.24e+7')} work</td>
                                </tr>
                              </tbody> 
                            </Table>
                            <h2>Chain graph</h2>

                            <VisNetworkReactComponent
                              data={networkData}
                              options={visOptions}
                              events={{click:this.chainGraphNodeClick}}
                            />
                            {r.detail.doubleSpends.length > 0 && <>
                            <h2>Double spends</h2>

                            {r.detail.doubleSpends.map((ds, i) => 
                            <Table striped className="doubleSpendDetail">
                              <thead>
                                <tr><th colspan="2">Double spend #{i}</th></tr>
                                <tr><th width="50%">Original</th><th width="50%">Replacement</th></tr>
                              </thead>
                              <tbody>
                                <tr>
                                  {ds.map((dsd, i) => <td key={i} align="center">TX <ShortHash hash={dsd.txHash} /><br/><small>(block <ShortHash hash={dsd.blockHash} />)</small></td>)}
                                </tr>
                                <tr>
                                  {ds.map((dsd, i) => {
                                  
                                  var rows = [];

                                  for(let j = 0; j < Math.max(dsd.in.length, dsd.out.length); j++) {

                                    let inCells = <><td>&nbsp;</td><td>&nbsp;</td></>;
                                    let outCells = <><td>&nbsp;</td><td>&nbsp;</td></>;
                                    
                                    if(dsd.in[j]) {
                                      inCells = <><td>{dsd.in[j].doubleSpent ? "Yes" : "No"}</td><td><ShortHash len={4} hash={dsd.in[j].prevoutTxID}/>/{dsd.in[j].prevoutIdx}</td></>;
                                    }

                                    if(dsd.out[j]) {
                                      outCells = <><td>{dsd.out[j].address}</td><td>{dsd.out[j].value}</td></>;
                                    }


                                    rows.push(<tr>{inCells}{outCells}</tr>);
                                  }
                                  
                                  
                                  return <td key={i} align="center">
                                    <Table striped className="doubleSpendTransaction">
                                      <thead>
                                        <tr><th colspan="2" width="50%">Inputs</th><th colspan="2" width="50%">Outputs</th></tr>
                                        <tr><th width="15%">Double spent</th><th width="35%">Outpoint</th><th width="35%">Script</th><th width="15%">Value</th></tr>
                                      </thead>
                                      <tbody>
                                        {rows}
                                      </tbody>
                                    </Table>
                                  </td>;
                                  
                                  })}
                                </tr>
                              </tbody>
                            </Table>)}
                            </>}
                          </Container>
                      }}/>
                      <Route exact path="/about">
                        <Container><Row><Col align="left">
                        <h1>About</h1>
                        <p>Reorg tracker analyzes consensus security of proof-of-work cryptocurrencies to provide empirical data on the rate of reorgs, how much fifty-one percent attacks cost and which coins are attackable in theory. The tracker actively observes over twenty cryptocurrency networks to detect reorgs and snapshots the Nicehash order book at regular intervals to provide an archive of hashrate rental market conditions.</p>
                        <p>Reorg tracker is made by:<RSNavLink href="https://dci.mit.edu/">MIT Digital Currency Initiative</RSNavLink></p>
                        <h2>Coin list</h2>
                        <p>Reorg tracker is currencly observing the following cryptocurrency networks:</p>
                        <Table>
                                <thead>
                                  <tr>
                                    <th>Ticker</th>
                                    <th>Name</th>
                                    <th>Tip hash</th>
                                    <th>Tip observed</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {this.state.coins.filter((c) => c.bestHash !== undefined && c.bestHash !== "").map((c) => <tr key={c.id}>
                                    <td>{c.ticker}</td>
                                    <td>{c.name}</td>
                                    <td>{c.bestHash.substring(0,8)}...{c.bestHash.substr(-8)}</td>
                                    <td><Moment fromNow ago>{c.bestHashObserved}</Moment> ago</td>
                                  </tr>)}
                                </tbody>
                              </Table>
                        </Col></Row></Container>
                      </Route>
                    </Switch>
                </Col>
              </Row>
            </Container>
            </Router>

    </div>
  );
  }
}

export default App;
