package gotest;

import com.chatserver.client.AesHelper;
import com.chatserver.client.ConvHelper;
import com.chatserver.client.PacketHelper;
import com.google.protobuf.InvalidProtocolBufferException;
import pb.*;

public class GoTest {
	public static void main(String[] args) {
		// 定义client，并连接
		com.chatserver.client.TcpHelper client = new com.chatserver.client.TcpHelper();
		boolean re = client.connect("115.29.174.190", 8989);
		if (!re) {
			System.out.println("connect error");
			return;
		}
		
		// 登录
		
		// 初始化登录包
		Pb.PbClientLogin.Builder clientLoginBuilder = Pb.PbClientLogin.newBuilder();
		clientLoginBuilder.setUuid("hello, 我是一个java程序");
		clientLoginBuilder.setVersion(3.14f);
		clientLoginBuilder.setTimestamp((int)(System.currentTimeMillis()/1000));
		
		// 发送登录包
		byte[] loginData = marshalPbClientLogin(clientLoginBuilder);
		re = client.send(loginData);
		if (!re) {
			System.out.println("send error");
			return;
		}
		
		// 读取数据
		byte[] readData = new byte[1024];
		int readSize = client.read(readData);
		if (-1 == readSize) {
			System.out.println("read error");
			return;
		}
		
		// 接收登录回执包
		// 假定读取的数据类型是登录回执包，同时假定readSize刚好是登录回执包的长度
		byte[] len = new byte[4];
		byte[] type = new byte[4];
		byte[] data = new byte[readSize-8];
		System.arraycopy(readData, 0, len, 0, 4);
		System.arraycopy(readData, 4, type, 0, 4);
		System.arraycopy(readData, 8, data, 0, readSize-8);
		
		// len 和 type直接读取
		System.out.println("len== " + ConvHelper.bytesToInt(len));
		System.out.println("type== " + ConvHelper.bytesToInt(type));
		if (ConvHelper.bytesToInt(type) == PacketHelper.PK_ServerAcceptLogin) {
			System.out.println("this is PK_ServerAcceptLogin");
		}
		
		// 解析登录回执包
		Pb.PbServerAcceptLogin serverAcceptLogin = unmarshalPbServerAcceptLogin(data);
		if (null == serverAcceptLogin) {
			System.out.println("unmarshalPbServerAcceptLogin error");
			return;
		}
		System.out.println(serverAcceptLogin.getLogin());
		System.out.println(serverAcceptLogin.getTipsMsg());
		System.out.println(serverAcceptLogin.getTimestamp());
		
		// 关闭连接
		client.close();
	}
	
	// PbClientLogin 
	// 将protobuf结构的builder 先序列化，再AES加密， 再打包
	public static byte[] marshalPbClientLogin(Pb.PbClientLogin.Builder b) {
		AesHelper aes = new AesHelper();
		Pb.PbClientLogin info = b.build();
		byte[] encrypted = aes.encrypt(info.toByteArray());
		return new PacketHelper(PacketHelper.PK_ClientLogin, encrypted).getBytes();
	}
	
	// PbServerAcceptLogin
	// data是在外面解包后传进来的，将data先解密，再反序列化为protobuf结构
	public static Pb.PbServerAcceptLogin unmarshalPbServerAcceptLogin(byte[] data) {
		AesHelper aes = new AesHelper();
		byte[] decrypted = aes.decrypt(data);
		
		Pb.PbServerAcceptLogin rev = null;
		try {
			rev = Pb.PbServerAcceptLogin.parseFrom(decrypted);
		} catch (InvalidProtocolBufferException e) {
			return null;
		}
		return rev;
	}
}
