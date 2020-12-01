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
      wrongWorkStart: new Date()-8*86400000,
      wrongWorkEnd: new Date()-86400000,
      transactionSetColors: [],
      coinTicker: "",
      coinId: -1,
    }

    for(var i = 0; i < 32; i++) {
      this.state.transactionSetColors.push('#' + Math.floor(Math.random()*16777215).toString(16));
    }
    this.toggle = this.toggle.bind(this);
    this.wrongWorkLater = this.wrongWorkLater.bind(this);
    this.wrongWorkEarlier = this.wrongWorkEarlier.bind(this);

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
          })
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
      this.setState({pools:pools, poolCoins: coins, locations : locations, coinId: coins[0].id})}
    );
    fetch("https://pooldetective.org/api/public/wrongwork/all").then(r=>r.json()).then((r)=>{
      this.setState({wrongwork:r.map((ww)=> {
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
                          This table shows the pools monitored by PoolDetective for the selected coin ({selectedCoin.name}). <br/>&nbsp;<br/>For each pool it shows the most recent work the pool sent us: both the time it was received, and the hash of the block it's building on top of.<br/>&nbsp;<br/> Under normal circumstances, the block we're building on should be the tip of the {selectedCoin.name} blockchain, which currently is: <code>{selectedCoin.bestHash}</code></p></center>
                        <Container>
                          <Row>
                            <Col>Transaction sets:</Col>
                          </Row>
                          <Row>
                            {transactionSets.map((ts, i) => 
                              <Col xs={3} style={{color:this.state.transactionSetColors[i]}}>
                                <ShortHash hash={ts} left={0} right={8} />
                              </Col>
                            )}
                          </Row>
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
                                                    <span style={jobTimeOk ? undefined : wrongStyle}>
                                                      <Moment format={jobTimeDate ? "L" : "H:mm:ss.SSS A"}>{new Date(observer.lastJobReceived)}</Moment>
                                                    </span>
                                                    <br/>
                                                    <small style={jobHashOk ? undefined : wrongStyle}>
                                                      <ShortHash left={0} right={8} hash={observer.lastJobPrevHash}></ShortHash>
                                                    </small>
                                                    <br/>
                                                    <small style={{color: this.state.transactionSetColors[transactionSets.indexOf(observer.lastJobMerkleProof)]}}>
                                                      <ShortHash left={0} right={8} hash={observer.lastJobMerkleProof}></ShortHash>
                                                    </small>
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
                                    <td>{pools.find((p) => p.id === ww.poolId).name}</td>
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
