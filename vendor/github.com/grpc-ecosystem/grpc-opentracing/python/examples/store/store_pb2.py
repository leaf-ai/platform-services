# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: store.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='store.proto',
  package='store',
  syntax='proto3',
  serialized_pb=_b('\n\x0bstore.proto\x12\x05store\"\x07\n\x05\x45mpty\"\x1e\n\x0e\x41\x64\x64ItemRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\"!\n\x11RemoveItemRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\",\n\x12RemoveItemResponse\x12\x16\n\x0ewas_successful\x18\x01 \x01(\x08\" \n\x10QueryItemRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\"/\n\x10QuantityResponse\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\r\n\x05\x63ount\x18\x02 \x01(\x05\x32\xc7\x03\n\x05Store\x12\x30\n\x07\x41\x64\x64Item\x12\x15.store.AddItemRequest\x1a\x0c.store.Empty\"\x00\x12\x33\n\x08\x41\x64\x64Items\x12\x15.store.AddItemRequest\x1a\x0c.store.Empty\"\x00(\x01\x12\x43\n\nRemoveItem\x12\x18.store.RemoveItemRequest\x1a\x19.store.RemoveItemResponse\"\x00\x12\x46\n\x0bRemoveItems\x12\x18.store.RemoveItemRequest\x1a\x19.store.RemoveItemResponse\"\x00(\x01\x12:\n\rListInventory\x12\x0c.store.Empty\x1a\x17.store.QuantityResponse\"\x00\x30\x01\x12\x43\n\rQueryQuantity\x12\x17.store.QueryItemRequest\x1a\x17.store.QuantityResponse\"\x00\x12I\n\x0fQueryQuantities\x12\x17.store.QueryItemRequest\x1a\x17.store.QuantityResponse\"\x00(\x01\x30\x01\x62\x06proto3')
)
_sym_db.RegisterFileDescriptor(DESCRIPTOR)




_EMPTY = _descriptor.Descriptor(
  name='Empty',
  full_name='store.Empty',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=22,
  serialized_end=29,
)


_ADDITEMREQUEST = _descriptor.Descriptor(
  name='AddItemRequest',
  full_name='store.AddItemRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='store.AddItemRequest.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=31,
  serialized_end=61,
)


_REMOVEITEMREQUEST = _descriptor.Descriptor(
  name='RemoveItemRequest',
  full_name='store.RemoveItemRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='store.RemoveItemRequest.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=63,
  serialized_end=96,
)


_REMOVEITEMRESPONSE = _descriptor.Descriptor(
  name='RemoveItemResponse',
  full_name='store.RemoveItemResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='was_successful', full_name='store.RemoveItemResponse.was_successful', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=98,
  serialized_end=142,
)


_QUERYITEMREQUEST = _descriptor.Descriptor(
  name='QueryItemRequest',
  full_name='store.QueryItemRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='store.QueryItemRequest.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=144,
  serialized_end=176,
)


_QUANTITYRESPONSE = _descriptor.Descriptor(
  name='QuantityResponse',
  full_name='store.QuantityResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='store.QuantityResponse.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='count', full_name='store.QuantityResponse.count', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=178,
  serialized_end=225,
)

DESCRIPTOR.message_types_by_name['Empty'] = _EMPTY
DESCRIPTOR.message_types_by_name['AddItemRequest'] = _ADDITEMREQUEST
DESCRIPTOR.message_types_by_name['RemoveItemRequest'] = _REMOVEITEMREQUEST
DESCRIPTOR.message_types_by_name['RemoveItemResponse'] = _REMOVEITEMRESPONSE
DESCRIPTOR.message_types_by_name['QueryItemRequest'] = _QUERYITEMREQUEST
DESCRIPTOR.message_types_by_name['QuantityResponse'] = _QUANTITYRESPONSE

Empty = _reflection.GeneratedProtocolMessageType('Empty', (_message.Message,), dict(
  DESCRIPTOR = _EMPTY,
  __module__ = 'store_pb2'
  # @@protoc_insertion_point(class_scope:store.Empty)
  ))
_sym_db.RegisterMessage(Empty)

AddItemRequest = _reflection.GeneratedProtocolMessageType('AddItemRequest', (_message.Message,), dict(
  DESCRIPTOR = _ADDITEMREQUEST,
  __module__ = 'store_pb2'
  # @@protoc_insertion_point(class_scope:store.AddItemRequest)
  ))
_sym_db.RegisterMessage(AddItemRequest)

RemoveItemRequest = _reflection.GeneratedProtocolMessageType('RemoveItemRequest', (_message.Message,), dict(
  DESCRIPTOR = _REMOVEITEMREQUEST,
  __module__ = 'store_pb2'
  # @@protoc_insertion_point(class_scope:store.RemoveItemRequest)
  ))
_sym_db.RegisterMessage(RemoveItemRequest)

RemoveItemResponse = _reflection.GeneratedProtocolMessageType('RemoveItemResponse', (_message.Message,), dict(
  DESCRIPTOR = _REMOVEITEMRESPONSE,
  __module__ = 'store_pb2'
  # @@protoc_insertion_point(class_scope:store.RemoveItemResponse)
  ))
_sym_db.RegisterMessage(RemoveItemResponse)

QueryItemRequest = _reflection.GeneratedProtocolMessageType('QueryItemRequest', (_message.Message,), dict(
  DESCRIPTOR = _QUERYITEMREQUEST,
  __module__ = 'store_pb2'
  # @@protoc_insertion_point(class_scope:store.QueryItemRequest)
  ))
_sym_db.RegisterMessage(QueryItemRequest)

QuantityResponse = _reflection.GeneratedProtocolMessageType('QuantityResponse', (_message.Message,), dict(
  DESCRIPTOR = _QUANTITYRESPONSE,
  __module__ = 'store_pb2'
  # @@protoc_insertion_point(class_scope:store.QuantityResponse)
  ))
_sym_db.RegisterMessage(QuantityResponse)


try:
  # THESE ELEMENTS WILL BE DEPRECATED.
  # Please use the generated *_pb2_grpc.py files instead.
  import grpc
  from grpc.framework.common import cardinality
  from grpc.framework.interfaces.face import utilities as face_utilities
  from grpc.beta import implementations as beta_implementations
  from grpc.beta import interfaces as beta_interfaces


  class StoreStub(object):

    def __init__(self, channel):
      """Constructor.

      Args:
        channel: A grpc.Channel.
      """
      self.AddItem = channel.unary_unary(
          '/store.Store/AddItem',
          request_serializer=AddItemRequest.SerializeToString,
          response_deserializer=Empty.FromString,
          )
      self.AddItems = channel.stream_unary(
          '/store.Store/AddItems',
          request_serializer=AddItemRequest.SerializeToString,
          response_deserializer=Empty.FromString,
          )
      self.RemoveItem = channel.unary_unary(
          '/store.Store/RemoveItem',
          request_serializer=RemoveItemRequest.SerializeToString,
          response_deserializer=RemoveItemResponse.FromString,
          )
      self.RemoveItems = channel.stream_unary(
          '/store.Store/RemoveItems',
          request_serializer=RemoveItemRequest.SerializeToString,
          response_deserializer=RemoveItemResponse.FromString,
          )
      self.ListInventory = channel.unary_stream(
          '/store.Store/ListInventory',
          request_serializer=Empty.SerializeToString,
          response_deserializer=QuantityResponse.FromString,
          )
      self.QueryQuantity = channel.unary_unary(
          '/store.Store/QueryQuantity',
          request_serializer=QueryItemRequest.SerializeToString,
          response_deserializer=QuantityResponse.FromString,
          )
      self.QueryQuantities = channel.stream_stream(
          '/store.Store/QueryQuantities',
          request_serializer=QueryItemRequest.SerializeToString,
          response_deserializer=QuantityResponse.FromString,
          )


  class StoreServicer(object):

    def AddItem(self, request, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')

    def AddItems(self, request_iterator, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')

    def RemoveItem(self, request, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')

    def RemoveItems(self, request_iterator, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')

    def ListInventory(self, request, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')

    def QueryQuantity(self, request, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')

    def QueryQuantities(self, request_iterator, context):
      context.set_code(grpc.StatusCode.UNIMPLEMENTED)
      context.set_details('Method not implemented!')
      raise NotImplementedError('Method not implemented!')


  def add_StoreServicer_to_server(servicer, server):
    rpc_method_handlers = {
        'AddItem': grpc.unary_unary_rpc_method_handler(
            servicer.AddItem,
            request_deserializer=AddItemRequest.FromString,
            response_serializer=Empty.SerializeToString,
        ),
        'AddItems': grpc.stream_unary_rpc_method_handler(
            servicer.AddItems,
            request_deserializer=AddItemRequest.FromString,
            response_serializer=Empty.SerializeToString,
        ),
        'RemoveItem': grpc.unary_unary_rpc_method_handler(
            servicer.RemoveItem,
            request_deserializer=RemoveItemRequest.FromString,
            response_serializer=RemoveItemResponse.SerializeToString,
        ),
        'RemoveItems': grpc.stream_unary_rpc_method_handler(
            servicer.RemoveItems,
            request_deserializer=RemoveItemRequest.FromString,
            response_serializer=RemoveItemResponse.SerializeToString,
        ),
        'ListInventory': grpc.unary_stream_rpc_method_handler(
            servicer.ListInventory,
            request_deserializer=Empty.FromString,
            response_serializer=QuantityResponse.SerializeToString,
        ),
        'QueryQuantity': grpc.unary_unary_rpc_method_handler(
            servicer.QueryQuantity,
            request_deserializer=QueryItemRequest.FromString,
            response_serializer=QuantityResponse.SerializeToString,
        ),
        'QueryQuantities': grpc.stream_stream_rpc_method_handler(
            servicer.QueryQuantities,
            request_deserializer=QueryItemRequest.FromString,
            response_serializer=QuantityResponse.SerializeToString,
        ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
        'store.Store', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


  class BetaStoreServicer(object):
    """The Beta API is deprecated for 0.15.0 and later.

    It is recommended to use the GA API (classes and functions in this
    file not marked beta) for all further purposes. This class was generated
    only to ease transition from grpcio<0.15.0 to grpcio>=0.15.0."""
    def AddItem(self, request, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)
    def AddItems(self, request_iterator, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)
    def RemoveItem(self, request, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)
    def RemoveItems(self, request_iterator, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)
    def ListInventory(self, request, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)
    def QueryQuantity(self, request, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)
    def QueryQuantities(self, request_iterator, context):
      context.code(beta_interfaces.StatusCode.UNIMPLEMENTED)


  class BetaStoreStub(object):
    """The Beta API is deprecated for 0.15.0 and later.

    It is recommended to use the GA API (classes and functions in this
    file not marked beta) for all further purposes. This class was generated
    only to ease transition from grpcio<0.15.0 to grpcio>=0.15.0."""
    def AddItem(self, request, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()
    AddItem.future = None
    def AddItems(self, request_iterator, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()
    AddItems.future = None
    def RemoveItem(self, request, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()
    RemoveItem.future = None
    def RemoveItems(self, request_iterator, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()
    RemoveItems.future = None
    def ListInventory(self, request, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()
    def QueryQuantity(self, request, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()
    QueryQuantity.future = None
    def QueryQuantities(self, request_iterator, timeout, metadata=None, with_call=False, protocol_options=None):
      raise NotImplementedError()


  def beta_create_Store_server(servicer, pool=None, pool_size=None, default_timeout=None, maximum_timeout=None):
    """The Beta API is deprecated for 0.15.0 and later.

    It is recommended to use the GA API (classes and functions in this
    file not marked beta) for all further purposes. This function was
    generated only to ease transition from grpcio<0.15.0 to grpcio>=0.15.0"""
    request_deserializers = {
      ('store.Store', 'AddItem'): AddItemRequest.FromString,
      ('store.Store', 'AddItems'): AddItemRequest.FromString,
      ('store.Store', 'ListInventory'): Empty.FromString,
      ('store.Store', 'QueryQuantities'): QueryItemRequest.FromString,
      ('store.Store', 'QueryQuantity'): QueryItemRequest.FromString,
      ('store.Store', 'RemoveItem'): RemoveItemRequest.FromString,
      ('store.Store', 'RemoveItems'): RemoveItemRequest.FromString,
    }
    response_serializers = {
      ('store.Store', 'AddItem'): Empty.SerializeToString,
      ('store.Store', 'AddItems'): Empty.SerializeToString,
      ('store.Store', 'ListInventory'): QuantityResponse.SerializeToString,
      ('store.Store', 'QueryQuantities'): QuantityResponse.SerializeToString,
      ('store.Store', 'QueryQuantity'): QuantityResponse.SerializeToString,
      ('store.Store', 'RemoveItem'): RemoveItemResponse.SerializeToString,
      ('store.Store', 'RemoveItems'): RemoveItemResponse.SerializeToString,
    }
    method_implementations = {
      ('store.Store', 'AddItem'): face_utilities.unary_unary_inline(servicer.AddItem),
      ('store.Store', 'AddItems'): face_utilities.stream_unary_inline(servicer.AddItems),
      ('store.Store', 'ListInventory'): face_utilities.unary_stream_inline(servicer.ListInventory),
      ('store.Store', 'QueryQuantities'): face_utilities.stream_stream_inline(servicer.QueryQuantities),
      ('store.Store', 'QueryQuantity'): face_utilities.unary_unary_inline(servicer.QueryQuantity),
      ('store.Store', 'RemoveItem'): face_utilities.unary_unary_inline(servicer.RemoveItem),
      ('store.Store', 'RemoveItems'): face_utilities.stream_unary_inline(servicer.RemoveItems),
    }
    server_options = beta_implementations.server_options(request_deserializers=request_deserializers, response_serializers=response_serializers, thread_pool=pool, thread_pool_size=pool_size, default_timeout=default_timeout, maximum_timeout=maximum_timeout)
    return beta_implementations.server(method_implementations, options=server_options)


  def beta_create_Store_stub(channel, host=None, metadata_transformer=None, pool=None, pool_size=None):
    """The Beta API is deprecated for 0.15.0 and later.

    It is recommended to use the GA API (classes and functions in this
    file not marked beta) for all further purposes. This function was
    generated only to ease transition from grpcio<0.15.0 to grpcio>=0.15.0"""
    request_serializers = {
      ('store.Store', 'AddItem'): AddItemRequest.SerializeToString,
      ('store.Store', 'AddItems'): AddItemRequest.SerializeToString,
      ('store.Store', 'ListInventory'): Empty.SerializeToString,
      ('store.Store', 'QueryQuantities'): QueryItemRequest.SerializeToString,
      ('store.Store', 'QueryQuantity'): QueryItemRequest.SerializeToString,
      ('store.Store', 'RemoveItem'): RemoveItemRequest.SerializeToString,
      ('store.Store', 'RemoveItems'): RemoveItemRequest.SerializeToString,
    }
    response_deserializers = {
      ('store.Store', 'AddItem'): Empty.FromString,
      ('store.Store', 'AddItems'): Empty.FromString,
      ('store.Store', 'ListInventory'): QuantityResponse.FromString,
      ('store.Store', 'QueryQuantities'): QuantityResponse.FromString,
      ('store.Store', 'QueryQuantity'): QuantityResponse.FromString,
      ('store.Store', 'RemoveItem'): RemoveItemResponse.FromString,
      ('store.Store', 'RemoveItems'): RemoveItemResponse.FromString,
    }
    cardinalities = {
      'AddItem': cardinality.Cardinality.UNARY_UNARY,
      'AddItems': cardinality.Cardinality.STREAM_UNARY,
      'ListInventory': cardinality.Cardinality.UNARY_STREAM,
      'QueryQuantities': cardinality.Cardinality.STREAM_STREAM,
      'QueryQuantity': cardinality.Cardinality.UNARY_UNARY,
      'RemoveItem': cardinality.Cardinality.UNARY_UNARY,
      'RemoveItems': cardinality.Cardinality.STREAM_UNARY,
    }
    stub_options = beta_implementations.stub_options(host=host, metadata_transformer=metadata_transformer, request_serializers=request_serializers, response_deserializers=response_deserializers, thread_pool=pool, thread_pool_size=pool_size)
    return beta_implementations.dynamic_stub(channel, 'store.Store', cardinalities, options=stub_options)
except ImportError:
  pass
# @@protoc_insertion_point(module_scope)
