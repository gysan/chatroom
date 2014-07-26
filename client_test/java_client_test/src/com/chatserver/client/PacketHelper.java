package com.chatserver.client;

//---------------------------数据包构造------------------------------------

//（1）字节序：大端模式

//（2）数据包组成：包长 + 类型 + 包体

//包长：4字节，int，整个数据包的长度

//类型：4字节，int

//包体：字节数组，byte[]

//包长和类型用明文传输，包体由结构体protobuf序列化后再进行AES加密得到。

//本类只负责打包

//--------------------------------------------------------------------------

public class PacketHelper {
	
	// 数据包的类型， 详见pb.proto
	public static int PK_ClientLogin = 1;
	public static int PK_ServerAcceptLogin = 2;
	public static int PK_ClientLogout = 3;
	public static int PK_ClientPing = 4;
	public static int PK_C2CTextChat = 5;
	
	// 数据包组成
	private int pacLen;
	private int pacType;
	private byte[] pacData;
	
	// type为包的类型， data为protobuf序列化后并已经加密了的数据
	public PacketHelper(int type, byte[] data) {
		this.pacData = data;
		this.pacLen = 8 + this.pacData.length;
		this.pacType = type;
	}
	
	public byte[] getBytes() {
		byte[] bytes = new byte[this.pacLen];
		System.arraycopy(ConvHelper.intToBytes(this.pacLen), 0, bytes, 0, 4);
		System.arraycopy(ConvHelper.intToBytes(this.pacType), 0, bytes, 4, 4);
		System.arraycopy(this.pacData, 0, bytes, 8, this.pacData.length);
		return bytes;
	}
	
	public static void main(String[] args) {
		// 测试
		PacketHelper ph = new PacketHelper(PacketHelper.PK_ClientLogin, "hello world, 哥在写java".getBytes());
		byte[] b = ph.getBytes();
		
		byte[] len = new byte[4];
		byte[] type = new byte[4];
		byte[] data = new byte[b.length-8];
		System.arraycopy(b, 0, len, 0, 4);
		System.arraycopy(b, 4, type, 0, 4);
		System.arraycopy(b, 8, data, 0, b.length-8);
		
		System.out.println(ConvHelper.bytesToInt(len));
		System.out.println(ConvHelper.bytesToInt(type));
		System.out.println(new String(data));
	}
}
