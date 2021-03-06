// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package verify

import (
	"errors"
	"fmt"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-sero/zero/txs/zstate"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

func CheckUint(i *utils.U256) bool {
	u := i.ToUint256()
	m := u[31] & (0xFF)
	if m != 0 {
		return false
	} else {
		return true
	}
}

func Verify(s *stx.T, state *zstate.ZState) (e error) {
	return Verify_state1(s, state)
}
func Verify_state1(s *stx.T, state *zstate.ZState) (e error) {

	t := utils.TR_enter("Miner-Verify-----Pre")

	balance_desc := cpt.BalanceDesc{}

	hash_z := s.ToHash_for_sign()
	balance_desc.Hash = hash_z

	if !CheckUint(&s.Fee.Value) {
		e = errors.New("txs.verify check fee too big")
		return
	}

	{
		asset_desc := cpt.AssetDesc{
			Tkn_currency: s.Fee.Currency,
			Tkn_value:    s.Fee.Value.ToUint256(),
			Tkt_category: keys.Empty_Uint256,
			Tkt_value:    keys.Empty_Uint256,
		}
		cpt.GenAssetCC(&asset_desc)
		balance_desc.Oout_accs = append(balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
	}

	if !keys.PKrValid(&s.From) {
		e = errors.New("txs.verify from is invalid")
		return
	}

	if !keys.VerifyPKr(&hash_z, &s.Sign, &s.From) {
		e = errors.New("txs.verify from verify failed")
		return
	}

	t.Renter("Miner-Verify-----o_ins")

	for _, in_o := range s.Desc_O.Ins {
		if ok := state.State.HasIn(&in_o.Root); ok {
			e = errors.New("txs.verify in already in nils")
			return
		} else {
		}
		if src, err := state.State.GetOut(&in_o.Root); e == nil {
			if src != nil {
				g := cpt.VerifyInputSDesc{}
				g.Ehash = hash_z
				g.Nil = in_o.Nil
				g.RootCM = *src.ToRootCM()
				g.Sign = in_o.Sign
				g.Pkr = *src.ToPKr()
				if err := cpt.VerifyInputS(&g); err != nil {
					e = errors.New("txs.Verify: in_o verify failed!")
					return
				} else {
					asset := src.Out_O.Asset.ToFlatAsset()
					asset_desc := cpt.AssetDesc{
						Tkn_currency: asset.Tkn.Currency,
						Tkn_value:    asset.Tkn.Value.ToUint256(),
						Tkt_category: asset.Tkt.Category,
						Tkt_value:    asset.Tkt.Value,
					}
					cpt.GenAssetCC(&asset_desc)
					balance_desc.Oin_accs = append(balance_desc.Oin_accs, asset_desc.Asset_cc[:]...)
				}
			} else {
				e = errors.New("txs.Verify: in_o not find in the outs!")
				return
			}
		} else {
			e = err
			return
		}
	}

	t.Renter("Miner-Verify-----o_outs")
	for _, out_o := range s.Desc_O.Outs {
		if out_o.Asset.Tkn != nil {
			if !CheckUint(&out_o.Asset.Tkn.Value) {
				e = errors.New("txs.verify check balance too big")
				return
			} else {
				{
					asset := out_o.Asset.ToFlatAsset()
					asset_desc := cpt.AssetDesc{
						Tkn_currency: asset.Tkn.Currency,
						Tkn_value:    asset.Tkn.Value.ToUint256(),
						Tkt_category: asset.Tkt.Category,
						Tkt_value:    asset.Tkt.Value,
					}
					cpt.GenAssetCC(&asset_desc)
					balance_desc.Oout_accs = append(balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
				}
			}
		}
	}

	t.Renter("Miner-Verify-----pkgs")
	if s.Desc_Pkg.Transfer != nil {
		if pg := state.Pkgs.GetPkg(&s.Desc_Pkg.Transfer.Id); pg == nil {
			e = fmt.Errorf("Can not find pkg of the id %v", hexutil.Encode(s.Desc_Pkg.Transfer.Id[:]))
			return
		} else {
			if keys.VerifyPKr(&hash_z, &s.Desc_Pkg.Transfer.Sign, &pg.Pack.PKr) {
			} else {
				e = fmt.Errorf("Can not verify pkg sign of the id %v", hexutil.Encode(s.Desc_Pkg.Transfer.Id[:]))
				return
			}
		}
	}

	if s.Desc_Pkg.Close != nil {
		if pg := state.Pkgs.GetPkg(&s.Desc_Pkg.Close.Id); pg == nil {
			e = fmt.Errorf("Can not find pkg of the id %v", hexutil.Encode(s.Desc_Pkg.Close.Id[:]))
			return
		} else {
			if keys.VerifyPKr(&hash_z, &s.Desc_Pkg.Close.Sign, &pg.Pack.PKr) {
				balance_desc.Zin_acms = append(balance_desc.Zin_acms, pg.Pack.Pkg.AssetCM[:]...)
			} else {
				e = fmt.Errorf("Can not verify pkg sign of the id %v", hexutil.Encode(s.Desc_Pkg.Close.Id[:]))
				return
			}
		}
	}

	t.Renter("Miner-Verify-----z_ins")
	for _, in_z := range s.Desc_Z.Ins {
		if ok := state.State.HasIn(&in_z.Nil); ok {
			e = errors.New("txs.verify in already in nils")
			return
		} else {
			if out, err := state.State.GetOut(&in_z.Anchor); err != nil {
				e = err
				return
			} else {
				if out == nil {
					e = errors.New("txs.verify can not find out for anchor")
				} else {
				}
			}
		}
	}

	t.Renter("Miner-Verify-----desc_zs")
	if err := verifyDesc_Zs(s, &balance_desc); err != nil {
		e = err
		return
	} else {
	}

	t.Renter("Miner-Verify-----balance_desc")
	balance_desc.Bcr = s.Bcr
	balance_desc.Bsign = s.Bsign
	if err := cpt.VerifyBalance(&balance_desc); err != nil {
		e = err
		return
	} else {
		t.Leave()
		return
	}
}
