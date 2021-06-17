pragma solidity ^0.6.0;

// SPDX-License-Identifier: UNLICENSED


contract relayerData{
    //address count 
    uint256 addressCount = 0;
    
    
    // data manager
    mapping(address => bool) private manager;
    address master;
    
    
    //data userinfo
    struct userInfo {
        address user;
        uint dayCount;
        uint daySign;
        uint256 amount;
    }
    
    modifier onlyManager() {
        require(manager[msg.sender],"onlyManager");
        _;
    }
    
    constructor() public {
        manager[msg.sender] = true;    
        master = msg.sender;
    }
    
    
}