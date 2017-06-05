//椭圆曲线加密原理参考：
//1、http://www.pediy.com/kssd/pediy06/pediy6014.htm
//2、http://www.hyperelliptic.org/EFD/g1p/auto-shortw.html
package main

import (
	"base58"
	"crypto/ecdsa"    //椭圆加密算法
	"crypto/elliptic" //椭圆曲线
	"crypto/rand"     //为随机数
	"math/big"        //大整数
)

//密钥对
type Keypair struct {
	Public  []byte //公钥，经Base58处理，对外公开，类似银行账号
	Private []byte //私钥，经Base58处理，自己保存，类似账号密码
	//Base58编码是可逆的，可以再回推出原来的字节。
	//Base58处理的好处：便于阅读，效率比较高。
}

//取随机数(伪随机数)，使用椭圆加密算法(Ep224)生成公钥和私钥对
//此处采用Go自带的Ep224曲线，比特币使用的是Ep256k1曲线，是对Ep256的改进版
func GenerateNewKeypair() *Keypair {
	//生成密钥对，计算过程大致如下：
	//1、选定一条椭圆曲线Ep(a,b)，并取椭圆曲线上一点，作为基点G(Gx,Gy)。
	//	1) p224.Gx, _ = new(big.Int).SetString("b70e0cbd6bb4bf7f321390b94a03c1d356c21122343280d6115c1d21", 16)
	//	2) p224.Gy, _ = new(big.Int).SetString("bd376388b5f723fb4c22dfe6cd4375a05a07476444d5819985007e34", 16)
	//2、选择一个私有密钥k（根据随机数rand.Reader生成k(pk.D)
	//3、根据K=kG生成公钥K(PublicKey.X,PublicKey.Y)
	kp, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)

	//将公钥的x和Y的值拼接成大整数，长度为56字节=28(KEY_SIZE)*2
	pb := bigJoin(26, kp.PublicKey.X, kp.PublicKey.Y)

	//使用base58编码
	public := base58.EncodeBig([]byte{}, pb)
	private := base58.EncodeBig([]byte{}, kp.D)

	pk := Keypair{Public: public, Private: private}
	return &pk
}

//使用私钥为哈希值签名,签名的目的是防止伪造
func (k *Keypair) Sign(hash []byte) ([]byte, error) {
	//解码Base58格式私钥
	priv, err := base58.DecodeToBig(k.Private)
	if err != nil {
		return nil, err
	}

	//解码Base58公钥私钥
	pub, err := base58.DecodeToBig(k.Public)
	if err != nil {
		return nil, err
	}
	//切分公钥获取x,y值
	pubb := splitBig(pub, 2)
	x, y := pubb[0], pubb[1]

	key := ecdsa.PrivateKey{ecdsa.PublicKey{elliptic.P224(), x, y}, priv}

	//使用私钥为哈希值签名
	r, s, err := ecdsa.Sign(rand.Reader, &key, hash)
	//将获取值r,s拼接，并转换为Base58格式
	return base58.EncodeBig([]byte{}, bigJoin(KEY_SIZE, r, s)), nil
}

//验证签名
func SignatureVerify(publicKey, sign, hash []byte) bool {
	//将公钥解码为大整数
	pub, _ := base58.DecodeToBig(publicKey)
	pubs := splitBig(pub, 2)
	x, y := pubs[0], pubs[1]
	//组建公钥
	public := ecdsa.PublicKey{elliptic.P224(), x, y}

	//切分sign
	s, _ := base58.DecodeToBig(sign)
	sl := splitBig(s, 2)
	r, s := sl[0], sl[1]
	//调用验证方法，并返回验证结果
	return ecdsa.Verify(&public, hash, r, s)
}

//将大整数安装固定长度拼接
func bigJoin(expectedLen int, bigs ...*big.Int) *big.Int {
	bs := []byte{}
	for i, b := range bigs {
		by := b.Bytes()
		dif := expectedLen - len(by)
		if dif > 0 && i != 0 {
			by = append(ArrayOfBytes(dif, '0'), by...)
		}
		bs = append(bs, by...)
	}
	return new(big.Int).SetBytes(bs)
}

//拆分大整数
func splitBig(b *big.Int, parts int) []*big.Int {
	bs := b.Bytes()
	if len(bs)%2 != 0 {
		bs = append([]byte{0}, bs...)
	}
	l := len(bs) / parts
	as := make([]*big.Int, parts)

	for i, _ := range as {
		as[i] = new(big.Int).SetBytes(bs[i*l : (i+1)*l])
	}

	return as
}
