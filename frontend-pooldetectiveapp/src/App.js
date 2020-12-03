import React from 'react';
import {
  BrowserRouter as Router,
  NavLink,
  Switch,
  Route
} from "react-router-dom";
import Moment from 'react-moment';
import {FaClock, FaWallet, FaBuromobelexperte} from 'react-icons/fa';
import './App.css';
import {Button, Form, FormGroup, Label, Input, Table, Container, Row, Col, NavLink as RSNavLink, Navbar, Nav, NavbarToggler, NavItem, NavbarBrand, Collapse} from 'reactstrap';
import * as numeral from 'numeral';

const ShortHash = (props) => {

  if(!props.hash)
  {
    return <></>
  }
  var len = props.len || 8;
  var left = props.left === undefined ? len : props.left;
  var right = props.right === undefined ? len : props.right;



  return <>{props.hash.substring(0,left)}...{props.hash.substr(0-right)}</>
}

const nameSort = ( a, b ) => {
  if ( a.name < b.name ){
    return -1;
  }
  if ( a.name > b.name ){
    return 1;
  }
  return 0;
}

const wrongStyle = {color:'#ff0000'};

const randDarkColor = function() {
  var lum = -0.25;
  var hex = String('#' + Math.random().toString(16).slice(2, 8).toUpperCase()).replace(/[^0-9a-f]/gi, '');
  if (hex.length < 6) {
      hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
  }
  var rgb = "#",
      c, i;
  for (i = 0; i < 3; i++) {
      c = parseInt(hex.substr(i * 2, 2), 16);
      c = Math.round(Math.min(Math.max(0, c + (c * lum)), 255)).toString(16);
      rgb += ("00" + c).substr(c.length);
  }
  return rgb;
}

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpen: false,
      pools: [],
      prevPools: [],
      coins: [],
      locations: [],
      poolCoins: [],
      wrongwork: [],
      emptyBlockWork : [],
      wrongWorkStart: new Date()-8*86400000,
      wrongWorkEnd: new Date()-86400000,
      emptyBlockWorkStart: new Date(),
      transactionSetColors: {},
      coinTicker: "",
      coinId: -1,
    }

    this.state.emptyBlockWorkStart = new Date(this.state.emptyBlockWorkStart.valueOf() - (this.state.emptyBlockWorkStart.valueOf()%86400000) - 86400000);
    this.toggle = this.toggle.bind(this);
    this.wrongWorkLater = this.wrongWorkLater.bind(this);
    this.wrongWorkEarlier = this.wrongWorkEarlier.bind(this);
    this.emptyBlockWorkLater = this.emptyBlockWorkLater.bind(this);
    this.emptyBlockWorkEarlier = this.emptyBlockWorkEarlier.bind(this);
    this.updateTransactionSetColors = this.updateTransactionSetColors.bind(this);

    this.ws = new WebSocket("wss://pooldetective.org/api/ws");
  
    this.ws.onmessage = (evt) => {
        var msg = JSON.parse(evt.data)
        if(msg.t === 'p') {
          this.setState((state) => {
            state.prevPools = Array.from(state.pools);
            for(let i = 0; i < state.pools.length; i++) {
              for(let j = 0; j < state.pools[i].observers.length; j++) {
                if(state.pools[i].observers[j].id === msg.m.id) {
                  clearTimeout(state.pools[i].observers[j].timeout)
                  state.pools[i].observers[j] = Object.assign({}, msg.m, {highlight: true, timeout: setTimeout(() => { this.setState((state) => { state.pools[i].observers[j] = Object.assign({}, state.pools[i].observers[j], {highlight: false}); return state; }); }, 2000)})
                }
              }
            }
          }, this.updateTransactionSetColors)
        } else if (msg.t === 'c') {
          this.setState((state) => {
            for(var i = 0; i < state.coins.length; i++) {
              if(state.coins[i].id === msg.m.id) {
                state.coins[i] = Object.assign({},msg.m);
                break;
              }
            }
            return state 
          })
        }
    }

    this.ws.onmessage = this.ws.onmessage.bind(this);
  }

 

  updateTransactionSetColors() {
    this.setState((s) => {
      s.pools.forEach((p) => {
        p.observers.forEach((o) => {
          if(!s.transactionSetColors[o.lastJobMerkleProof]) {
            s.transactionSetColors[o.lastJobMerkleProof] = randDarkColor();
          }
        })
      })
      return s;
    });
  }
  
  wrongWorkLater() {
    let wrongWorkStart = this.state.wrongWorkStart+7*86400000
    if(wrongWorkStart > new Date()-8*86400000) {
      wrongWorkStart = new Date()-8*86400000
    }

    let wrongWorkEnd = this.state.wrongWorkEnd+7*86400000
    if(wrongWorkEnd > new Date()-1*86400000) {
      wrongWorkEnd = new Date()-1*86400000
    }
    this.setState({wrongWorkStart,wrongWorkEnd})
  }

  wrongWorkEarlier() {
    let wrongWorkStart = this.state.wrongWorkStart-7*86400000
    let wrongWorkEnd = this.state.wrongWorkEnd-7*86400000
    this.setState({wrongWorkStart,wrongWorkEnd})
  }

  emptyBlockWorkLater() {
    let emptyBlockWorkStart = this.state.emptyBlockWorkStart+86400000
    if(emptyBlockWorkStart > new Date()-86400000) {
      emptyBlockWorkStart = new Date()-86400000
    }


    this.setState({emptyBlockWorkStart})
  }

  emptyBlockWorkEarlier() {
    let emptyBlockWorkStart = this.state.emptyBlockWorkStart-86400000
    this.setState({emptyBlockWorkStart})
  }

  componentDidMount() {
    fetch("https://pooldetective.org/api/public/coins").then(r=>r.json()).then((r)=>{
      this.setState({coins:r.sort(nameSort)})
    });
    fetch("https://pooldetective.org/api/public/pools").then(r=>r.json()).then((r)=>{
      var coins = [];
      var locations = [];
      var pools = r.sort(nameSort)
      pools.forEach((p) => {
        if(!coins.find((c) => c.id === p.coinId)) {
          coins.push({id:p.coinId, name:p.coinName});
        }
        p.observers.forEach((o) => {
          if(!locations.find((l) => l.id === o.locationId)) {
            locations.push({id:o.locationId, name:o.locationName});
          }
        })
      })
      coins = coins.sort(nameSort)
      locations = locations.sort(nameSort)
      this.setState({pools:pools, poolCoins: coins, locations : locations, coinId: coins[0].id}, this.updateTransactionSetColors)}
    );
    fetch("https://pooldetective.org/api/public/wrongwork/all").then(r=>r.json()).then((r)=>{
      this.setState({wrongwork:r.map((ww)=> {
        ww.observedOn = new Date(ww.observedOn);
        return ww;
      })})
    });
    fetch("https://pooldetective.org/api/public/emptyblockwork/all").then(r=>r.json()).then((r)=>{
      this.setState({emptyBlockWork:r.map((ww)=> {
        ww.observedOn = new Date(ww.observedOn);
        return ww;
      })})
    });
  }

  toggle() {
    this.setState({isOpen:!this.state.isOpen})
  } 

  render() {

    const selectedCoin = this.state.coins.find((c) => (c.id === parseInt(this.state.coinId))) || {};
    const pools = this.state.pools.filter((p) => p.coinId == this.state.coinId);
    var transactionSets = [];
    pools.forEach((p) => {
      p.observers.forEach((o) => {
        if(transactionSets.indexOf(o.lastJobMerkleProof) === -1) {
          transactionSets.push(o.lastJobMerkleProof);
        }
      })
    })
    transactionSets = transactionSets.sort((a,b) => (a > b) ? 1 : ((b > a) ? -1 : 0));
    const locations = this.state.locations.filter((l) => pools.find((p) => p.observers.find((o) => o.locationId === l.id)));
    const prevPools = this.state.prevPools.filter((p) => pools.find((po) => p.id === po.id));

  return (
    <div className="App">
      <Router>
      <Navbar fixed color="light" light expand="md">
        <NavbarBrand href="/">Pool Detective</NavbarBrand>
        <NavbarToggler onClick={this.toggle} />
        <Collapse isOpen={this.state.isOpen} navbar>
          <Nav className="mr-auto" navbar>
            <NavItem>
              <RSNavLink tag={NavLink} to="/">Current Pool Work</RSNavLink>
            </NavItem>
            <NavItem>
              <RSNavLink tag={NavLink} to="/history">Wrong Pool Work</RSNavLink>
            </NavItem>
            <NavItem>
              <RSNavLink tag={NavLink} to="/empty">Empty Blocks</RSNavLink>
            </NavItem>
            <NavItem> 
              <RSNavLink tag={Button} color="link" href="https://pooldetective.org/">About</RSNavLink>
            </NavItem>
         
            
            </Nav></Collapse>
            <Nav>
              <NavItem style={{width: 300}}>
                <Row>
                  <Label xs={4} for="coinTicker">Coin:</Label>
                  <Col  xs={8}>
                  <Input type="select" name="coinTicker" value={this.state.coinId} onChange={(e) => { this.setState({coinId : e.target.value}); }}>
                  {this.state.coins.filter((c) => this.state.poolCoins.find(pc => pc.id === c.id)).map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                  </Input>
                  </Col>
                  </Row>
                </NavItem>
                <NavItem style={{width: 300}}>
                  <Container>
                  <Row><Label xs={12}>
                    Tip: <ShortHash left={0} right={8} hash={selectedCoin.bestHash}></ShortHash> (<Moment format="LTS">{selectedCoin.bestHashObserved}</Moment>)
                  </Label></Row>
                  </Container>
                </NavItem>
            </Nav></Navbar>
            <Container fluid>
              <Row>
                <Col>
                    <Switch>
                      <Route exact path="/" render={({history}) => (
                      <Container><Row><Col align="left">
                        <center><h1>Current Pool Work</h1>
                        <p>
                          This table shows the pools monitored by PoolDetective for the selected coin ({selectedCoin.name}). <br/>&nbsp;<br/>For each pool it shows the most recent work the pool sent us: both the time it was received, and the hash of the block it's building on top of.<br/>&nbsp;<br/> Under normal circumstances, the block we're building on should be the tip of the {selectedCoin.name} blockchain, which currently is: <code>{selectedCoin.bestHash}</code>
                        </p>
                         <Container>
                          <Row>
                            <Col xs={4}><FaClock /> - The time the last work from this pool was received</Col>
                            <Col xs={4}><FaBuromobelexperte /> - The hash of the block this pool building on</Col>
                            <Col xs={4}><FaWallet /> - The hash of the transaction set included in this pool's new block</Col>
                          </Row>
                        </Container>
                        </center>
                        <Container>
                          <Row>
                            <Col>
                              <Table striped>
                                <thead>
                                  <tr>
                                    <th style={{width:'20%'}}>Pool</th>
                                    {locations.map((l) => <th style={{textAlign:'center', width:'20%'}}>{l.name}</th>)}
                                  </tr>
                                </thead>
                                <tbody>
                                  {pools.map((r) => {
                                    var prev = prevPools.find((pp) => pp.id === r.id);

                                    return <tr key={r.id}>
                                      <td style={{verticalAlign:'middle'}}>{r.name}</td>
                                      {locations.map((l) => {
                                        
                                        var observer = r.observers.find((o) => o.locationId === l.id);
                                        if(observer) {
                                          const jobHashOk = (observer.lastJobPrevHash === selectedCoin.bestHash);
                                          const jobTimeOk = ((new Date().valueOf() - new Date(observer.lastJobReceived).valueOf()) < 300000);
                                          const jobTimeDate = ((new Date().valueOf() - new Date(observer.lastJobReceived).valueOf()) > 86400000);
                                          let highlight = observer.highlight;

                                          return <td align="center" className={`${jobHashOk&&jobTimeOk ? 'good' : 'bad'}${highlight ? 'Highlight' : ''}`}>
                                                    <div className="time" style={jobTimeOk ? undefined : wrongStyle}>
                                                      <FaClock /> <Moment format={jobTimeDate ? "L" : "h:mm:ss.SSS A"}>{new Date(observer.lastJobReceived)}</Moment>
                                                    </div>
                                                    <div className="hashes">
                                                    <div className="workHashPill" title="The block hash this pool is building on" style={jobHashOk ? undefined : wrongStyle}>
                                                      <FaBuromobelexperte /> <ShortHash left={0} right={6} hash={observer.lastJobPrevHash}></ShortHash>
                                                    </div>
                                                    <div className="workHashPill" title="The transaction set this pool is trying to include in its next block"  style={{color: 'white', backgroundColor: this.state.transactionSetColors[observer.lastJobMerkleProof]}}>
                                                      <FaWallet /> <ShortHash left={0} right={6} hash={observer.lastJobMerkleProof}></ShortHash>
                                                    </div>
                                                    </div>
                                                  </td>
                                        } else {
                                          return <td>&nbsp;</td>
                                        }
                                      
                                      })}
                                      
                                  </tr>;
                                })}
                                </tbody>
                              </Table>
                            </Col>
                          </Row>
                        </Container>
                        </Col></Row></Container>)} />
                        <Route exact path="/history" render={({history}) => (
                          <Container><Row><Col align="left">
                            <center><h1>Wrong Pool Work</h1>
                              <p>
                                This table shows when the pools we monitor sent us unexpected work. This is work that builds on top of a previous block that we matched to a different blockchain.
                                <br/>&nbsp;<br/>
                              </p>

                              <p>
                                <Container>
                                  <Row>
                                    <Col xs={2}><Button color="link" onClick={this.wrongWorkEarlier}>Earlier</Button></Col>
                                    <Col><Moment format="LL">{this.state.wrongWorkStart}</Moment> - <Moment format="LL">{this.state.wrongWorkEnd}</Moment></Col>
                                    <Col xs={2}><Button color="link" onClick={this.wrongWorkLater}>Later</Button></Col>
                                  </Row>
                                </Container>
                              </p>
                            </center>
                            <Container>
                              <Row>
                                <Col>
                                  <Table striped>
                                    <thead>
                                      <tr>
                                        <th>Date</th>
                                        <th>Pool</th>
                                        <th>Location</th>
                                        <th>Expected coin</th>
                                        <th>Got coin</th>
                                        <th>Wrong Jobs</th>
                                        <th>Time spent</th>
                                      </tr>
                                    </thead>
                                    <tbody>
                                      {this.state.wrongwork.filter((ww) => ww.observedOn.valueOf() >= this.state.wrongWorkStart.valueOf() && ww.observedOn.valueOf() < this.state.wrongWorkEnd.valueOf() && ww.expectedCoinId === selectedCoin.id).map((ww) => <tr>
                                        <td><Moment format="L">{ww.observedOn}</Moment></td>
                                        <td>{this.state.pools.find((p) => p.id === ww.poolId).name}</td>
                                        <td>{ww.location}</td>
                                        <td>{ww.expectedCoinName}</td>
                                        <td>{ww.wrongCoinName}</td>
                                        <td>{ww.wrongJobs} ({numeral(ww.wrongJobs/ww.totalJobs).format('%')})</td>
                                        <td>{ww.wrongTimeMs/1000} seconds ({numeral(ww.wrongTimeMs/ww.totalTimeMs).format('%')})</td>
                                      </tr>)}
                                    </tbody>
                                  </Table>
                                </Col>
                              </Row>
                            </Container>
                            </Col></Row></Container>)} />
                            <Route exact path="/empty" render={({history}) => (
                          <Container><Row><Col align="left">
                            <center><h1>Empty Block Work</h1>
                              <p>
                                This table shows how often the pools we monitor sent us work for mining an empty block. This is expected to happen after finding a new block, but should roughly be the same for all pools.
                                <br/>&nbsp;<br/>
                              </p>

                              <p>
                                <Container>
                                  <Row>
                                    <Col xs={2}><Button color="link" onClick={this.emptyBlockWorkEarlier}>Earlier</Button></Col>
                                    <Col><Moment format="LL">{this.state.emptyBlockWorkStart}</Moment></Col>
                                    <Col xs={2}><Button color="link" onClick={this.emptyBlockWorkLater}>Later</Button></Col>
                                  </Row>
                                </Container>
                              </p>
                            </center>
                            <Container>
                              <Row>
                                <Col>
                                  <Table striped>
                                    <thead>
                                      <tr>
                                        <th>Date</th>
                                        <th>Pool</th>
                                        <th>Location</th>
                                        <th>Coin</th>
                                        <th>Empty Block Jobs</th>
                                        <th>Empty Block time spent</th>
                                      </tr>
                                    </thead>
                                    <tbody>
                                      {this.state.emptyBlockWork.filter((ebw) => pools.find((p) => p.id === ebw.poolId) && ebw.observedOn.valueOf() >= this.state.emptyBlockWorkStart.valueOf() && ebw.observedOn.valueOf() < this.state.emptyBlockWorkStart.valueOf()+86400000 && ebw.coinId === selectedCoin.id).map((ebw) => <tr>
                                        <td><Moment format="L">{ebw.observedOn}</Moment></td>
                                        <td>{pools.find((p) => p.id === ebw.poolId).name}</td>
                                        <td>{ebw.location}</td>
                                        <td>{ebw.coinName}</td>
                                        <td>{ebw.emptyBlockJobs} ({numeral(ebw.emptyBlockJobs/ebw.totalJobs).format('%')})</td>
                                        <td>{ebw.emptyBlockTimeMs/1000} seconds ({numeral(ebw.emptyBlockTimeMs/ebw.totalTimeMs).format('%')})</td>
                                      </tr>)}
                                    </tbody>
                                  </Table>
                                </Col>
                              </Row>
                            </Container>
                            </Col></Row></Container>)} />
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
