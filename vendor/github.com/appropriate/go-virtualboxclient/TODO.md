* Handle `<vbox:InvalidObjectFault>` when new login is required
* Call `IManagedObjectRef::release` and/or `IWebsessionManager::logoff` as necessary to release object references and avoid leaks
