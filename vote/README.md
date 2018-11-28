#### 投票智能合约

> 先看官方给的例子：
>
> ```
> peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mycc -v 1.0 -c 
> '{"Args":["init","a","100","b","200"]}'
> ```
>
> 初始化a，a有100块；初始化b，b有200块。
>
> ```
> peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n mycc -c '{"Args":["invoke","a","b","10"]}'
> ```
>
> a给b10块钱。
>
> 最大的问题是：**这个函数，在这个联盟中的成员都可以调用。**
>
> 那么，如何判断a是a？b是b？
>
> 这个问题衍生出来：如何判定区块链中的私有个人数据与共有数据？
>
> 这里提供一种简单的解决思路。

##### 前提

1. 能调用此智能合约 -> 调用者一定是此区块联盟中的用户并用合法的公私调用
2. 此用户属于该联盟并有相应权限 -> 此用户有权调用智能合约
3. chaincode提供stub.GetCreator()方法得到调用者的公钥
4. 区块存取的数据的结构是map

##### 结论

* 把用户的公钥作为key，用户的私有数据作为value。需要私有数据，直接从stub.GetCreator()中拿出key。用户的公钥就像这个联盟中天生的身份证。

详细代码请看源码。

#### 用法：

chaincode的部署者即组织的admin，他作为这个投票环境的推动人，推动的方法为next，投票有三个阶段：

1. 开始：联盟中的用户在此投票阶段注册（register），比如名字，是投票者还是竞选者？
2. 投票：通过getCandidates方法得到竞选者信息，得到他的key，通过vote方法，将票投给竞选者
3. 清算：每个人都可以通过balance方法得到投票结果



另外，通过history方法可以得到相应key对应的所有操作

