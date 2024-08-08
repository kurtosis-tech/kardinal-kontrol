# Changelog

## [0.1.11](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.10...0.1.11) (2024-08-08)


### Features

* Postgres DB integration with ORM ([#144](https://github.com/kurtosis-tech/kardinal-kontrol/issues/144)) ([13cdd1c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/13cdd1c0b72943a3dfa20d76c28616c8f2fb729a))
* update some of the hardcoded values to represent the demo flow ([#146](https://github.com/kurtosis-tech/kardinal-kontrol/issues/146)) ([bb98cd4](https://github.com/kurtosis-tech/kardinal-kontrol/commit/bb98cd463c8e481472a0f56056d7da4dec677c27))

## [0.1.10](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.9...0.1.10) (2024-08-06)


### Features

* handle external services v1 ([#111](https://github.com/kurtosis-tech/kardinal-kontrol/issues/111)) ([295654a](https://github.com/kurtosis-tech/kardinal-kontrol/commit/295654abb6b45c95eaa1e40882e58498788b6d31))


### Bug Fixes

* fix frontend needlessly re-rendering the graph ([#143](https://github.com/kurtosis-tech/kardinal-kontrol/issues/143)) ([30d1580](https://github.com/kurtosis-tech/kardinal-kontrol/commit/30d1580283562624e1108c998ed8fb3338415039))
* Sort topology nodes and edges for deterministic response ([#142](https://github.com/kurtosis-tech/kardinal-kontrol/issues/142)) ([6704ca6](https://github.com/kurtosis-tech/kardinal-kontrol/commit/6704ca62296ba82f36a90c9b1ff44f93a2a07e55))

## [0.1.9](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.8...0.1.9) (2024-08-03)


### Bug Fixes

* apply config to the right action ([#137](https://github.com/kurtosis-tech/kardinal-kontrol/issues/137)) ([61a7e5b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/61a7e5b593eee1a4e0168d4ef1f08e6cd79c2c67))

## [0.1.8](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.7...0.1.8) (2024-08-03)


### Bug Fixes

* ci again ([#135](https://github.com/kurtosis-tech/kardinal-kontrol/issues/135)) ([3164162](https://github.com/kurtosis-tech/kardinal-kontrol/commit/3164162d951cda7b0d6f4b9011a97f576d3baaea))

## [0.1.7](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.6...0.1.7) (2024-08-03)


### Bug Fixes

* add release output ([#133](https://github.com/kurtosis-tech/kardinal-kontrol/issues/133)) ([12933fa](https://github.com/kurtosis-tech/kardinal-kontrol/commit/12933faf9990458f01eaf66cb3d32d62332288ed))

## [0.1.6](https://github.com/kurtosis-tech/kardinal-kontrol/compare/v0.1.6...0.1.6) (2024-08-02)


### Features

* add boilerplate for frontend app ([6e5986a](https://github.com/kurtosis-tech/kardinal-kontrol/commit/6e5986ac7292339b274478a1c83d0c4528fda4d4))
* add generalized cluster topology ([#98](https://github.com/kurtosis-tech/kardinal-kontrol/issues/98)) ([2cb6948](https://github.com/kurtosis-tech/kardinal-kontrol/commit/2cb69489075b7705df4f8a09320af7f80c10691f))
* add generic topology data structure ([#95](https://github.com/kurtosis-tech/kardinal-kontrol/issues/95)) ([f2953bf](https://github.com/kurtosis-tech/kardinal-kontrol/commit/f2953bf11d68cb51df64e246c983ea791a6ece4d))
* add postgres icon ([#78](https://github.com/kurtosis-tech/kardinal-kontrol/issues/78)) ([99206bb](https://github.com/kurtosis-tech/kardinal-kontrol/commit/99206bb5409ccd87983438db827a9f10a8a5677f))
* add support for a plugin system ([#83](https://github.com/kurtosis-tech/kardinal-kontrol/issues/83)) ([2eee456](https://github.com/kurtosis-tech/kardinal-kontrol/commit/2eee4560ea2401c1637ab645fe73e5d88b9949dc))
* add support for neon postgresql ([#53](https://github.com/kurtosis-tech/kardinal-kontrol/issues/53)) ([4a7288b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/4a7288b89f2067556c80dab6f9a77daaf2ce4625))
* Add the associated destination rule for the http route subset ([#106](https://github.com/kurtosis-tech/kardinal-kontrol/issues/106)) ([a0d0af6](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a0d0af6faa1e1d245ce2f9aa43ecaef628c41b09))
* adding instance tag for Kardinal EKS node group ([#105](https://github.com/kurtosis-tech/kardinal-kontrol/issues/105)) ([d5eb6f3](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d5eb6f397fe7903ddc6138b897c0d8cf482fa8bd))
* create envoy filters and authorization policies for tracing ([#107](https://github.com/kurtosis-tech/kardinal-kontrol/issues/107)) ([a46d82d](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a46d82d79de1926b0928e245c21c7db21cfb6960))
* do not run CI on draft PRs ([#100](https://github.com/kurtosis-tech/kardinal-kontrol/issues/100)) ([b10b26b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/b10b26bcd70ceccfbbfb789fcd2a4954692ac3b2))
* first draft of Kontrol plane UI ([#99](https://github.com/kurtosis-tech/kardinal-kontrol/issues/99)) ([98a212b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/98a212bf5de17bdb71adb8a0771c35dd88d5c953))
* for a stateful non http service duplicate its parent and add the right routing rule ([#113](https://github.com/kurtosis-tech/kardinal-kontrol/issues/113)) ([460e479](https://github.com/kurtosis-tech/kardinal-kontrol/commit/460e4792666e19cdd3e70dfcef73c7bb3c43ca6d))
* Generic deploy to prod flow ([#101](https://github.com/kurtosis-tech/kardinal-kontrol/issues/101)) ([070da14](https://github.com/kurtosis-tech/kardinal-kontrol/commit/070da14bf058f1142caf62b6680b84d8c44da769))
* graph styling update for demo flow ([#97](https://github.com/kurtosis-tech/kardinal-kontrol/issues/97)) ([ce66d85](https://github.com/kurtosis-tech/kardinal-kontrol/commit/ce66d85bf48ddecc5a58cf4f4f7a7b92d3fb70c9))
* improvements to cli ([#112](https://github.com/kurtosis-tech/kardinal-kontrol/issues/112)) ([fcd9353](https://github.com/kurtosis-tech/kardinal-kontrol/commit/fcd9353d53ec66723b26596899574b612f440e84))
* kardinal kluster deployment ([e5d4f5d](https://github.com/kurtosis-tech/kardinal-kontrol/commit/e5d4f5d7f4d66698c3ec3834bd0320e7913ae737))
* kardinal kluster deployment ([a48808c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a48808c640cf48aa581961215d923293af1be66d))
* Kardinal Kontrol cloudformation for the Kardinal and the Kardinal load balancer stacks ([#87](https://github.com/kurtosis-tech/kardinal-kontrol/issues/87)) ([183140c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/183140c73a2e96f7219d62b514e59251237f7b53))
* kontrol plane dashboard components ([#86](https://github.com/kurtosis-tech/kardinal-kontrol/issues/86)) ([fbcde84](https://github.com/kurtosis-tech/kardinal-kontrol/commit/fbcde846fcb3169f2c5f6e7a401ef7835e464831))
* make the graph look nicer for larger obd demo ([#128](https://github.com/kurtosis-tech/kardinal-kontrol/issues/128)) ([a284306](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a2843069d391b7ef26cebdfdefd437f6309c3d3e))
* merge flows ([#110](https://github.com/kurtosis-tech/kardinal-kontrol/issues/110)) ([2adc5dc](https://github.com/kurtosis-tech/kardinal-kontrol/commit/2adc5dc4754814a85d93b888a6ff1bcfd8ba44c2))
* new graph styling ([#85](https://github.com/kurtosis-tech/kardinal-kontrol/issues/85)) ([639bdf5](https://github.com/kurtosis-tech/kardinal-kontrol/commit/639bdf576c4bd3f3f2ae48cc760b140121034af2))
* only publish and deploy on releases ([#104](https://github.com/kurtosis-tech/kardinal-kontrol/issues/104)) ([e1f418e](https://github.com/kurtosis-tech/kardinal-kontrol/commit/e1f418eef7a4abd8322912a506106da73e56af8a))
* parallelize CI ([#102](https://github.com/kurtosis-tech/kardinal-kontrol/issues/102)) ([fe0fd27](https://github.com/kurtosis-tech/kardinal-kontrol/commit/fe0fd27c358e8d18d09c4746b42bff9acf9ebd0b))
* remove docs voting app demo ([#80](https://github.com/kurtosis-tech/kardinal-kontrol/issues/80)) ([d3d732f](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d3d732fe751f530b2018b824d2090874d3a9d816))
* set up storybook frontend devtool ([#81](https://github.com/kurtosis-tech/kardinal-kontrol/issues/81)) ([0c0df9e](https://github.com/kurtosis-tech/kardinal-kontrol/commit/0c0df9edf053cbf8e155b047d749937d3cf6ae5a))
* Support k8s service / deployment configs sent by the CLI ([#79](https://github.com/kurtosis-tech/kardinal-kontrol/issues/79)) ([1e4b7b4](https://github.com/kurtosis-tech/kardinal-kontrol/commit/1e4b7b44714807fbd372c89cb46e2a5768c06525))
* Update topology request to include flows information ([#122](https://github.com/kurtosis-tech/kardinal-kontrol/issues/122)) ([868cf46](https://github.com/kurtosis-tech/kardinal-kontrol/commit/868cf4621317a212e2b4cac8e7a592a179eab2db))


### Bug Fixes

* added ca cert package ([#74](https://github.com/kurtosis-tech/kardinal-kontrol/issues/74)) ([7cb0915](https://github.com/kurtosis-tech/kardinal-kontrol/commit/7cb091591420539f2ada6d201664766ec5478c56))
* another fix for releases ([#117](https://github.com/kurtosis-tech/kardinal-kontrol/issues/117)) ([5a88fc4](https://github.com/kurtosis-tech/kardinal-kontrol/commit/5a88fc4c6d4e585c593ab17be02f07ddafc70706))
* aws login during CI ([#119](https://github.com/kurtosis-tech/kardinal-kontrol/issues/119)) ([d2321b6](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d2321b650e5530f288d491125fd23c6db265d527))
* Deep copy k8s service config to get a real copy ([#84](https://github.com/kurtosis-tech/kardinal-kontrol/issues/84)) ([45b362f](https://github.com/kurtosis-tech/kardinal-kontrol/commit/45b362fde3fdb9b30cc8c7ec188c738a5e5b7879))
* deploy depends on publish ([#103](https://github.com/kurtosis-tech/kardinal-kontrol/issues/103)) ([7fa0c47](https://github.com/kurtosis-tech/kardinal-kontrol/commit/7fa0c476ebb49db0d508b60faffda4ebb50234c5))
* fix some of the animation bugs in the Cytoscape graph ([#75](https://github.com/kurtosis-tech/kardinal-kontrol/issues/75)) ([c6d7b13](https://github.com/kurtosis-tech/kardinal-kontrol/commit/c6d7b13bf03525183de90ffe552d1decb8375b19))
* fix the release process ([#115](https://github.com/kurtosis-tech/kardinal-kontrol/issues/115)) ([87ed709](https://github.com/kurtosis-tech/kardinal-kontrol/commit/87ed709c36c95fb7095747990368a6e36e218e2c))
* make plugins use file io for returning results ([#88](https://github.com/kurtosis-tech/kardinal-kontrol/issues/88)) ([dbeea12](https://github.com/kurtosis-tech/kardinal-kontrol/commit/dbeea1209d17d18eebfc0ec81ee526a58a2256b2))
* nix run kontrol-service  ([#125](https://github.com/kurtosis-tech/kardinal-kontrol/issues/125)) ([c71f273](https://github.com/kurtosis-tech/kardinal-kontrol/commit/c71f2732a028fd6a512192e549a9fb566c120978))
* nix sandbox during release ([#126](https://github.com/kurtosis-tech/kardinal-kontrol/issues/126)) ([8d3be57](https://github.com/kurtosis-tech/kardinal-kontrol/commit/8d3be57699005c98b44efd63d86f28026b7a5e9f))
* release please to release latest deployment changes ([#93](https://github.com/kurtosis-tech/kardinal-kontrol/issues/93)) ([22fe943](https://github.com/kurtosis-tech/kardinal-kontrol/commit/22fe943ad4d7904eedaef77e8742d6c01ce79422))
* remove demo from workspace ([#77](https://github.com/kurtosis-tech/kardinal-kontrol/issues/77)) ([9b91f9c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/9b91f9ca79822ca672bd875f10b699b33a0a1254))
* remove frontend requirement ([#73](https://github.com/kurtosis-tech/kardinal-kontrol/issues/73)) ([d958a0c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d958a0c64fa79ad574eaa0351e565d3def5bddb2))
* remove unneeded node dep ([e1f9705](https://github.com/kurtosis-tech/kardinal-kontrol/commit/e1f97059266e1dcf0d5197b8692b19444fa882b6))
* revert [#115](https://github.com/kurtosis-tech/kardinal-kontrol/issues/115) ([#130](https://github.com/kurtosis-tech/kardinal-kontrol/issues/130)) ([69c3fc6](https://github.com/kurtosis-tech/kardinal-kontrol/commit/69c3fc62b7376639dbff2c8463bfe4b9fbcee30b))
* slow down animation speed ([#96](https://github.com/kurtosis-tech/kardinal-kontrol/issues/96)) ([6321418](https://github.com/kurtosis-tech/kardinal-kontrol/commit/63214188f6c9090f6cd61c295707311bbfa248e5))
* update frontend hash on x86 ([#131](https://github.com/kurtosis-tech/kardinal-kontrol/issues/131)) ([bf8bd27](https://github.com/kurtosis-tech/kardinal-kontrol/commit/bf8bd273a21461c456ca7d27936d1917c53ecf9c))
* use printf instead of fatalf ([#72](https://github.com/kurtosis-tech/kardinal-kontrol/issues/72)) ([4712caf](https://github.com/kurtosis-tech/kardinal-kontrol/commit/4712cafcdcaf1b5f6a37d02ebb1a077171a75d2b))


### Miscellaneous Chores

* release 0.1.0 ([93f0c73](https://github.com/kurtosis-tech/kardinal-kontrol/commit/93f0c7342709c07e83ffec6351b40bc2d707d696))
* release 0.1.6 ([a9deaea](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a9deaea168b10bf6295a1031a01519b8fbbaccc1))

## [0.1.6](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.5...0.1.6) (2024-08-02)


### Bug Fixes

* nix sandbox during release ([#126](https://github.com/kurtosis-tech/kardinal-kontrol/issues/126)) ([8d3be57](https://github.com/kurtosis-tech/kardinal-kontrol/commit/8d3be57699005c98b44efd63d86f28026b7a5e9f))

## [0.1.5](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.4...0.1.5) (2024-08-02)


### Features

* first draft of Kontrol plane UI ([#99](https://github.com/kurtosis-tech/kardinal-kontrol/issues/99)) ([98a212b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/98a212bf5de17bdb71adb8a0771c35dd88d5c953))
* Update topology request to include flows information ([#122](https://github.com/kurtosis-tech/kardinal-kontrol/issues/122)) ([868cf46](https://github.com/kurtosis-tech/kardinal-kontrol/commit/868cf4621317a212e2b4cac8e7a592a179eab2db))


### Bug Fixes

* nix run kontrol-service  ([#125](https://github.com/kurtosis-tech/kardinal-kontrol/issues/125)) ([c71f273](https://github.com/kurtosis-tech/kardinal-kontrol/commit/c71f2732a028fd6a512192e549a9fb566c120978))

## [0.1.4](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.3...0.1.4) (2024-08-02)


### Bug Fixes

* aws login during CI ([#119](https://github.com/kurtosis-tech/kardinal-kontrol/issues/119)) ([d2321b6](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d2321b650e5530f288d491125fd23c6db265d527))

## [0.1.3](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.2...0.1.3) (2024-08-02)


### Bug Fixes

* another fix for releases ([#117](https://github.com/kurtosis-tech/kardinal-kontrol/issues/117)) ([5a88fc4](https://github.com/kurtosis-tech/kardinal-kontrol/commit/5a88fc4c6d4e585c593ab17be02f07ddafc70706))

## [0.1.2](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.1...0.1.2) (2024-08-02)


### Bug Fixes

* fix the release process ([#115](https://github.com/kurtosis-tech/kardinal-kontrol/issues/115)) ([87ed709](https://github.com/kurtosis-tech/kardinal-kontrol/commit/87ed709c36c95fb7095747990368a6e36e218e2c))

## [0.1.1](https://github.com/kurtosis-tech/kardinal-kontrol/compare/0.1.0...0.1.1) (2024-08-02)


### Features

* add generalized cluster topology ([#98](https://github.com/kurtosis-tech/kardinal-kontrol/issues/98)) ([2cb6948](https://github.com/kurtosis-tech/kardinal-kontrol/commit/2cb69489075b7705df4f8a09320af7f80c10691f))
* add generic topology data structure ([#95](https://github.com/kurtosis-tech/kardinal-kontrol/issues/95)) ([f2953bf](https://github.com/kurtosis-tech/kardinal-kontrol/commit/f2953bf11d68cb51df64e246c983ea791a6ece4d))
* Add the associated destination rule for the http route subset ([#106](https://github.com/kurtosis-tech/kardinal-kontrol/issues/106)) ([a0d0af6](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a0d0af6faa1e1d245ce2f9aa43ecaef628c41b09))
* adding instance tag for Kardinal EKS node group ([#105](https://github.com/kurtosis-tech/kardinal-kontrol/issues/105)) ([d5eb6f3](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d5eb6f397fe7903ddc6138b897c0d8cf482fa8bd))
* create envoy filters and authorization policies for tracing ([#107](https://github.com/kurtosis-tech/kardinal-kontrol/issues/107)) ([a46d82d](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a46d82d79de1926b0928e245c21c7db21cfb6960))
* do not run CI on draft PRs ([#100](https://github.com/kurtosis-tech/kardinal-kontrol/issues/100)) ([b10b26b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/b10b26bcd70ceccfbbfb789fcd2a4954692ac3b2))
* for a stateful non http service duplicate its parent and add the right routing rule ([#113](https://github.com/kurtosis-tech/kardinal-kontrol/issues/113)) ([460e479](https://github.com/kurtosis-tech/kardinal-kontrol/commit/460e4792666e19cdd3e70dfcef73c7bb3c43ca6d))
* Generic deploy to prod flow ([#101](https://github.com/kurtosis-tech/kardinal-kontrol/issues/101)) ([070da14](https://github.com/kurtosis-tech/kardinal-kontrol/commit/070da14bf058f1142caf62b6680b84d8c44da769))
* graph styling update for demo flow ([#97](https://github.com/kurtosis-tech/kardinal-kontrol/issues/97)) ([ce66d85](https://github.com/kurtosis-tech/kardinal-kontrol/commit/ce66d85bf48ddecc5a58cf4f4f7a7b92d3fb70c9))
* improvements to cli ([#112](https://github.com/kurtosis-tech/kardinal-kontrol/issues/112)) ([fcd9353](https://github.com/kurtosis-tech/kardinal-kontrol/commit/fcd9353d53ec66723b26596899574b612f440e84))
* kontrol plane dashboard components ([#86](https://github.com/kurtosis-tech/kardinal-kontrol/issues/86)) ([fbcde84](https://github.com/kurtosis-tech/kardinal-kontrol/commit/fbcde846fcb3169f2c5f6e7a401ef7835e464831))
* merge flows ([#110](https://github.com/kurtosis-tech/kardinal-kontrol/issues/110)) ([2adc5dc](https://github.com/kurtosis-tech/kardinal-kontrol/commit/2adc5dc4754814a85d93b888a6ff1bcfd8ba44c2))
* only publish and deploy on releases ([#104](https://github.com/kurtosis-tech/kardinal-kontrol/issues/104)) ([e1f418e](https://github.com/kurtosis-tech/kardinal-kontrol/commit/e1f418eef7a4abd8322912a506106da73e56af8a))
* parallelize CI ([#102](https://github.com/kurtosis-tech/kardinal-kontrol/issues/102)) ([fe0fd27](https://github.com/kurtosis-tech/kardinal-kontrol/commit/fe0fd27c358e8d18d09c4746b42bff9acf9ebd0b))


### Bug Fixes

* deploy depends on publish ([#103](https://github.com/kurtosis-tech/kardinal-kontrol/issues/103)) ([7fa0c47](https://github.com/kurtosis-tech/kardinal-kontrol/commit/7fa0c476ebb49db0d508b60faffda4ebb50234c5))
* release please to release latest deployment changes ([#93](https://github.com/kurtosis-tech/kardinal-kontrol/issues/93)) ([22fe943](https://github.com/kurtosis-tech/kardinal-kontrol/commit/22fe943ad4d7904eedaef77e8742d6c01ce79422))
* slow down animation speed ([#96](https://github.com/kurtosis-tech/kardinal-kontrol/issues/96)) ([6321418](https://github.com/kurtosis-tech/kardinal-kontrol/commit/63214188f6c9090f6cd61c295707311bbfa248e5))

## 0.1.0 (2024-07-16)


### Features

* add boilerplate for frontend app ([6e5986a](https://github.com/kurtosis-tech/kardinal-kontrol/commit/6e5986ac7292339b274478a1c83d0c4528fda4d4))
* add postgres icon ([#78](https://github.com/kurtosis-tech/kardinal-kontrol/issues/78)) ([99206bb](https://github.com/kurtosis-tech/kardinal-kontrol/commit/99206bb5409ccd87983438db827a9f10a8a5677f))
* add support for a plugin system ([#83](https://github.com/kurtosis-tech/kardinal-kontrol/issues/83)) ([2eee456](https://github.com/kurtosis-tech/kardinal-kontrol/commit/2eee4560ea2401c1637ab645fe73e5d88b9949dc))
* add support for neon postgresql ([#53](https://github.com/kurtosis-tech/kardinal-kontrol/issues/53)) ([4a7288b](https://github.com/kurtosis-tech/kardinal-kontrol/commit/4a7288b89f2067556c80dab6f9a77daaf2ce4625))
* kardinal kluster deployment ([e5d4f5d](https://github.com/kurtosis-tech/kardinal-kontrol/commit/e5d4f5d7f4d66698c3ec3834bd0320e7913ae737))
* kardinal kluster deployment ([a48808c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/a48808c640cf48aa581961215d923293af1be66d))
* new graph styling ([#85](https://github.com/kurtosis-tech/kardinal-kontrol/issues/85)) ([639bdf5](https://github.com/kurtosis-tech/kardinal-kontrol/commit/639bdf576c4bd3f3f2ae48cc760b140121034af2))
* remove docs voting app demo ([#80](https://github.com/kurtosis-tech/kardinal-kontrol/issues/80)) ([d3d732f](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d3d732fe751f530b2018b824d2090874d3a9d816))
* set up storybook frontend devtool ([#81](https://github.com/kurtosis-tech/kardinal-kontrol/issues/81)) ([0c0df9e](https://github.com/kurtosis-tech/kardinal-kontrol/commit/0c0df9edf053cbf8e155b047d749937d3cf6ae5a))
* Support k8s service / deployment configs sent by the CLI ([#79](https://github.com/kurtosis-tech/kardinal-kontrol/issues/79)) ([1e4b7b4](https://github.com/kurtosis-tech/kardinal-kontrol/commit/1e4b7b44714807fbd372c89cb46e2a5768c06525))


### Bug Fixes

* added ca cert package ([#74](https://github.com/kurtosis-tech/kardinal-kontrol/issues/74)) ([7cb0915](https://github.com/kurtosis-tech/kardinal-kontrol/commit/7cb091591420539f2ada6d201664766ec5478c56))
* Deep copy k8s service config to get a real copy ([#84](https://github.com/kurtosis-tech/kardinal-kontrol/issues/84)) ([45b362f](https://github.com/kurtosis-tech/kardinal-kontrol/commit/45b362fde3fdb9b30cc8c7ec188c738a5e5b7879))
* fix some of the animation bugs in the Cytoscape graph ([#75](https://github.com/kurtosis-tech/kardinal-kontrol/issues/75)) ([c6d7b13](https://github.com/kurtosis-tech/kardinal-kontrol/commit/c6d7b13bf03525183de90ffe552d1decb8375b19))
* make plugins use file io for returning results ([#88](https://github.com/kurtosis-tech/kardinal-kontrol/issues/88)) ([dbeea12](https://github.com/kurtosis-tech/kardinal-kontrol/commit/dbeea1209d17d18eebfc0ec81ee526a58a2256b2))
* remove demo from workspace ([#77](https://github.com/kurtosis-tech/kardinal-kontrol/issues/77)) ([9b91f9c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/9b91f9ca79822ca672bd875f10b699b33a0a1254))
* remove frontend requirement ([#73](https://github.com/kurtosis-tech/kardinal-kontrol/issues/73)) ([d958a0c](https://github.com/kurtosis-tech/kardinal-kontrol/commit/d958a0c64fa79ad574eaa0351e565d3def5bddb2))
* remove unneeded node dep ([e1f9705](https://github.com/kurtosis-tech/kardinal-kontrol/commit/e1f97059266e1dcf0d5197b8692b19444fa882b6))
* use printf instead of fatalf ([#72](https://github.com/kurtosis-tech/kardinal-kontrol/issues/72)) ([4712caf](https://github.com/kurtosis-tech/kardinal-kontrol/commit/4712cafcdcaf1b5f6a37d02ebb1a077171a75d2b))


### Miscellaneous Chores

* release 0.1.0 ([93f0c73](https://github.com/kurtosis-tech/kardinal-kontrol/commit/93f0c7342709c07e83ffec6351b40bc2d707d696))
