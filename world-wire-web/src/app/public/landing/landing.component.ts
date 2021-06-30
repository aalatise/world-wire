import {
  Component,
  OnInit,
  HostBinding
} from '@angular/core';


@Component({
  templateUrl: './landing.component.html',
  styleUrls: ['./landing.component.scss']
})
export class LandingComponent implements OnInit {

  @HostBinding('attr.class') cls = 'flex-fill';

  ngOnInit() { }

}
