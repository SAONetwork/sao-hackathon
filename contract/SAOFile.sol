// SPDX-License-Identifier: MIT

pragma solidity ^0.8.1;
import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";


contract SAOFile is ERC721, Ownable, ReentrancyGuard {

    using SafeMath for uint256;

    string private _base_uri;
    address public admin;
    uint256 public idx = 0;
    uint256 public orderIdx= 0;
    uint256 public fee = 1;
    uint256 public totalFee = 0;
    mapping(uint256 => uint256) public listing;
    mapping(uint256 => mapping(address => bool)) public buyer;
    mapping(address => uint256) public balances;

    struct Order {
        uint256 tokenId;
        address buyer;
        address seller;
        uint256 price;
        uint256 status;
        uint256 fee;
    }
    
    mapping(uint256 => Order) public orders;

    event Bought(uint256 indexed tokenId, address indexed buyer, uint256 indexed orderId, uint256 price, uint256 timestamp);
    event Listing(uint256 indexed tokenId, uint256 fileId, uint256 price, uint256 timestamp);
    event ChangePrice(uint256 indexed tokenId, uint256 price, uint256 timestamp);
    event Withdraw(address indexed user, uint256 amount, uint256 timestamp);
    event Download(uint256 indexed orderId, uint256 timestamp);

    constructor(string memory name, string memory symbol, string memory _uri) ERC721(name, symbol) {
        _base_uri = _uri;
    }

    function _baseURI() internal view virtual override returns (string memory) {
        return  _base_uri;
    }

    function setBaseURI(string memory _uri) external onlyOwner {
        _base_uri = _uri;
    }

    function mint(uint256 file_id, uint256 price) external {
        idx += 1;
        _mint(msg.sender, idx);
        listing[idx] = price;
        emit Listing(idx, file_id, price, block.timestamp);
    }

    function changePrice(uint256 tokenId, uint256 price) external {
        require(ownerOf(tokenId) == msg.sender, "not your nft");
        listing[tokenId] = price;
        emit ChangePrice(tokenId, price, block.timestamp);
    }

    function buy(uint256 tokenId) external payable {
        require(listing[tokenId] > 0, "not listing yet");
        require(msg.value == listing[tokenId], "payment amount error");
        require(buyer[tokenId][msg.sender] == false, "already bought");
        uint256 sale_fee = listing[tokenId].mul(fee).mul(100).div(10000);
        address owner = ownerOf(tokenId);
        Order memory order;
        order.tokenId = tokenId;
        order.buyer = msg.sender;
        order.seller = owner;
        order.price = listing[tokenId];
        order.fee = sale_fee;
        order.status = 1;
        orders[orderIdx] = order;
        buyer[tokenId][msg.sender] = true;
        emit Bought(tokenId, msg.sender, orderIdx, msg.value, block.timestamp);
        orderIdx++;
    }

    function finish(uint256 orderId) external onlyOwner {
        Order memory order = orders[orderId];
        require(order.status == 1, "error order status");

        totalFee += order.fee;
        uint256 tokenId = order.tokenId;
        address owner = ownerOf(tokenId);
        balances[owner] += order.price.sub(order.fee);
        buyer[tokenId][msg.sender] = true;
        delete orders[orderId];
        emit Download(orderId, block.timestamp);
    }

    function withdraw() external nonReentrant {
        if (msg.sender == admin) {
            payable(msg.sender).transfer(totalFee);
            emit Withdraw(msg.sender, totalFee, block.timestamp);
            totalFee = 0;
        } else {
            uint256 balance = balances[msg.sender];
            require(balance > 0, "insufficient balance");
            balances[msg.sender] = 0;
            payable(msg.sender).transfer(balance);
            emit Withdraw(msg.sender, balance, block.timestamp);
        }
    } 
}
